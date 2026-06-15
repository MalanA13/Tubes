// Package functional berisi functional test (integration test) end-to-end.
//
// Menggunakan SQLite in-memory — tidak perlu Docker, tidak perlu server eksternal.
// Setiap test mendapat database FRESH (via setupTest) sehingga terisolasi.
//
// Juga menggunakan stub TrackingClient dan OrderClient agar tidak butuh
// network ke microservice lain, tapi tetap menguji alur service + repository.
//
// Flow utama yang ditest:
//   CREATED → ScanIn(IN_HUB) → ScanOut(IN_TRANSIT)
//           → AssignCourier(OUT_DELIVERY) → UpdateDeliveryStatus(DELIVERED)
package functional

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tubes-cc/logistics/internal/courier"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/hub"
	sqliterepo "github.com/tubes-cc/logistics/infrastructure/sqlite"
)

// ================================================================
// STUB CLIENTS — pengganti HTTP client agar functional test
// tidak membutuhkan network ke microservice lain.
// ================================================================

// stubTrackingClient menyimpan semua event yang diterima ke slice in-memory.
// Digunakan untuk memverifikasi bahwa event dikirim dengan benar.
type stubTrackingClient struct {
    repo   *sqliterepo.ShipmentRepository
    events []*domain.TrackingEvent
}

func (s *stubTrackingClient) AddTrackingEvent(ctx context.Context, e *domain.TrackingEvent) error {
    s.events = append(s.events, e)

    // 🔥 INI KUNCI NYA
    return s.repo.AddTrackingEvent(ctx, e)
}

// stubOrderClient selalu menganggap resi valid (return nil).
// Untuk test "resi tidak valid", gunakan stubOrderClientInvalid.
type stubOrderClient struct{}

func (s *stubOrderClient) ValidateResi(_ context.Context, _ string) error { return nil }

// stubOrderClientInvalid selalu menolak resi.
type stubOrderClientInvalid struct{}

func (s *stubOrderClientInvalid) ValidateResi(_ context.Context, _ string) error {
	return domain.ErrResiInvalid
}

// ================================================================
// SETUP HELPER
// ================================================================

// testEnv menampung semua komponen yang dibutuhkan setiap test.
type testEnv struct {
	repo           *sqliterepo.ShipmentRepository
	trackingStub   *stubTrackingClient
	hubService     *hub.Service
	courierService *courier.Service
}

// setupTest membuat environment test yang segar (fresh DB in-memory).
// Dipanggil di awal setiap test agar terisolasi satu sama lain.
func setupTest(t *testing.T) *testEnv {
	t.Helper()

	// Buka SQLite in-memory — ":memory:" berarti hanya hidup selama koneksi terbuka
	db, err := sqliterepo.Open(":memory:")
	require.NoError(t, err)

	repo, err := sqliterepo.NewShipmentRepository(db)
	require.NoError(t, err)

	// Tutup DB saat test selesai
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	trackingStub := &stubTrackingClient{
		repo: repo,
	}	
	
	orderStub := &stubOrderClient{}

	return &testEnv{
		repo:           repo,
		trackingStub:   trackingStub,
		hubService:     hub.NewService(repo, trackingStub, orderStub),
		courierService: courier.NewService(repo, trackingStub, orderStub),
	}
}

// ================================================================
// FUNCTIONAL TESTS
// ================================================================

// TestFullDeliveryFlow_HappyPath menguji alur pengiriman lengkap dan sukses:
//   CREATED → IN_HUB → IN_TRANSIT → OUT_DELIVERY → DELIVERED
func TestFullDeliveryFlow_HappyPath(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	// --- Step 1: ScanIn ---
	t.Log("▶ Step 1: ScanIn di Hub Jakarta")
	err := env.hubService.ScanIn(ctx, "RESI-001", "HUB-JKT-01")
	require.NoError(t, err)

	s, err := env.repo.GetShipment(ctx, "RESI-001")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusInHub, s.Status)
	assert.Equal(t, "HUB-JKT-01", s.HubID)

	// --- Step 2: ScanOut ---
	t.Log("▶ Step 2: ScanOut dari Hub Jakarta")
	err = env.hubService.ScanOut(ctx, "RESI-001", "HUB-JKT-01")
	require.NoError(t, err)

	s, _ = env.repo.GetShipment(ctx, "RESI-001")
	assert.Equal(t, domain.StatusInTransit, s.Status)

	// --- Step 3: AssignCourier ---
	t.Log("▶ Step 3: Assign Kurir")
	err = env.courierService.AssignCourier(ctx, "RESI-001", "KURIR-BUDI-01")
	require.NoError(t, err)

	s, _ = env.repo.GetShipment(ctx, "RESI-001")
	assert.Equal(t, domain.StatusOutDelivery, s.Status)
	assert.Equal(t, "KURIR-BUDI-01", s.CourierID)

	// --- Step 4: UpdateDeliveryStatus → DELIVERED ---
	t.Log("▶ Step 4: Update Status DELIVERED")
	proofURL := "https://storage.example.com/proof/RESI-001.jpg"
	err = env.courierService.UpdateDeliveryStatus(ctx, "RESI-001", domain.StatusDelivered, proofURL)
	require.NoError(t, err)

	s, _ = env.repo.GetShipment(ctx, "RESI-001")
	assert.Equal(t, domain.StatusDelivered, s.Status)
	assert.Equal(t, proofURL, s.ProofURL)

	// --- Verifikasi tracking events ---
	events, err := env.repo.GetTrackingEvents(ctx, "RESI-001")
	require.NoError(t, err)
	require.Len(t, events, 4, "harus ada 4 tracking events")

	// Verifikasi setiap event ID unik (UUID — tidak ada duplikat)
	ids := map[string]bool{}
	for _, e := range events {
		assert.NotEmpty(t, e.ID, "event ID tidak boleh kosong")
		assert.False(t, ids[e.ID], "event ID harus unik, ditemukan duplikat: %s", e.ID)
		ids[e.ID] = true
	}

	// Verifikasi urutan status kronologis
	assert.Equal(t, domain.StatusInHub, events[0].Status)
	assert.Equal(t, domain.StatusInTransit, events[1].Status)
	assert.Equal(t, domain.StatusOutDelivery, events[2].Status)
	assert.Equal(t, domain.StatusDelivered, events[3].Status)

	// Verifikasi stub tracking client juga menerima 4 event
	assert.Len(t, env.trackingStub.events, 4)

	t.Log("✅ Happy path berhasil!")
}

// TestFailedDeliveryFlow menguji alur pengiriman yang gagal:
//   CREATED → IN_HUB → IN_TRANSIT → OUT_DELIVERY → FAILED
func TestFailedDeliveryFlow(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	require.NoError(t, env.hubService.ScanIn(ctx, "RESI-002", "HUB-BDG-01"))
	require.NoError(t, env.hubService.ScanOut(ctx, "RESI-002", "HUB-BDG-01"))
	require.NoError(t, env.courierService.AssignCourier(ctx, "RESI-002", "KURIR-ANI-01"))

	// FAILED tidak perlu proofURL
	err := env.courierService.UpdateDeliveryStatus(ctx, "RESI-002", domain.StatusFailed, "")
	require.NoError(t, err)

	s, _ := env.repo.GetShipment(ctx, "RESI-002")
	assert.Equal(t, domain.StatusFailed, s.Status)

	t.Log("✅ Failed delivery flow berhasil!")
}

// TestMultiHubTransit menguji paket melewati beberapa hub:
//   ScanIn(HUB-A) → ScanOut(HUB-A) → ScanIn(HUB-B) → ScanOut(HUB-B) → ...
func TestMultiHubTransit(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	// Hub 1: Jakarta
	require.NoError(t, env.hubService.ScanIn(ctx, "RESI-003", "HUB-JKT-01"))
	require.NoError(t, env.hubService.ScanOut(ctx, "RESI-003", "HUB-JKT-01"))

	// Cek status setelah keluar dari Hub JKT
	s, _ := env.repo.GetShipment(ctx, "RESI-003")
	assert.Equal(t, domain.StatusInTransit, s.Status)

	// Hub 2: Bandung — scan-in dari status IN_TRANSIT juga diizinkan
	require.NoError(t, env.hubService.ScanIn(ctx, "RESI-003", "HUB-BDG-01"))

	s, _ = env.repo.GetShipment(ctx, "RESI-003")
	assert.Equal(t, domain.StatusInHub, s.Status)
	assert.Equal(t, "HUB-BDG-01", s.HubID)

	require.NoError(t, env.hubService.ScanOut(ctx, "RESI-003", "HUB-BDG-01"))

	// Kirim ke penerima
	require.NoError(t, env.courierService.AssignCourier(ctx, "RESI-003", "KURIR-DENI-01"))
	require.NoError(t, env.courierService.UpdateDeliveryStatus(
		ctx, "RESI-003", domain.StatusDelivered, "https://proof.example.com/003.jpg",
	))

	// 6 events total: ScanIn(JKT), ScanOut(JKT), ScanIn(BDG), ScanOut(BDG), Assign, Delivered
	events, _ := env.repo.GetTrackingEvents(ctx, "RESI-003")
	assert.Len(t, events, 6)

	// Semua UUID unik
	ids := map[string]bool{}
	for _, e := range events {
		assert.False(t, ids[e.ID], "UUID duplikat ditemukan: %s", e.ID)
		ids[e.ID] = true
	}

	t.Log("✅ Multi-hub transit berhasil!")
}

// TestDuplicateScanIn_ShouldFail memastikan scan-in berulang di hub yang sama ditolak.
func TestDuplicateScanIn_ShouldFail(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	// Pertama kali: berhasil
	require.NoError(t, env.hubService.ScanIn(ctx, "RESI-004", "HUB-JKT-01"))

	// Kedua kali di hub yang sama: harus error
	err := env.hubService.ScanIn(ctx, "RESI-004", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrAlreadyScannedIn)

	t.Log("✅ Duplicate scan-in berhasil diblok!")
}

// TestScanOutBeforeScanIn_ShouldFail memastikan scan-out tanpa scan-in ditolak.
func TestScanOutBeforeScanIn_ShouldFail(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	// Buat shipment tapi belum scan-in
	require.NoError(t, env.repo.CreateShipmentIfNotExists(ctx, "RESI-005"))

	// Langsung scan-out — harus error karena status masih CREATED
	err := env.hubService.ScanOut(ctx, "RESI-005", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrInvalidStatus)

	t.Log("✅ Scan-out sebelum scan-in berhasil diblok!")
}

// TestAssignCourierBeforeScanOut_ShouldFail memastikan assign kurir sebelum scan-out ditolak.
func TestAssignCourierBeforeScanOut_ShouldFail(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	// Hanya scan-in, belum scan-out
	require.NoError(t, env.hubService.ScanIn(ctx, "RESI-006", "HUB-JKT-01"))

	// Assign kurir — harus error karena status IN_HUB, bukan IN_TRANSIT
	err := env.courierService.AssignCourier(ctx, "RESI-006", "KURIR-001")
	assert.ErrorIs(t, err, domain.ErrInvalidStatus)

	t.Log("✅ Assign kurir sebelum scan-out berhasil diblok!")
}

// TestDeliveredWithoutProof_ShouldFail memastikan DELIVERED tanpa proofURL ditolak.
func TestDeliveredWithoutProof_ShouldFail(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	require.NoError(t, env.hubService.ScanIn(ctx, "RESI-007", "HUB-SBY-01"))
	require.NoError(t, env.hubService.ScanOut(ctx, "RESI-007", "HUB-SBY-01"))
	require.NoError(t, env.courierService.AssignCourier(ctx, "RESI-007", "KURIR-EKO-01"))

	// DELIVERED tanpa proof URL — harus error
	err := env.courierService.UpdateDeliveryStatus(ctx, "RESI-007", domain.StatusDelivered, "")
	assert.ErrorIs(t, err, domain.ErrInvalidProofURL)

	t.Log("✅ DELIVERED tanpa proof URL berhasil diblok!")
}

// TestResiValidation_ShouldFail menguji bahwa resi tidak valid ditolak di semua operasi.
func TestResiValidation_ShouldFail(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	// Ganti dengan order client yang selalu menolak
	invalidOrderClient := &stubOrderClientInvalid{}
	hubSvc := hub.NewService(env.repo, env.trackingStub, invalidOrderClient)
	courierSvc := courier.NewService(env.repo, env.trackingStub, invalidOrderClient)

	err := hubSvc.ScanIn(ctx, "RESI-PALSU", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrResiInvalid)

	err = hubSvc.ScanOut(ctx, "RESI-PALSU", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrResiInvalid)

	err = courierSvc.AssignCourier(ctx, "RESI-PALSU", "KURIR-001")
	assert.ErrorIs(t, err, domain.ErrResiInvalid)

	err = courierSvc.UpdateDeliveryStatus(ctx, "RESI-PALSU", domain.StatusFailed, "")
	assert.ErrorIs(t, err, domain.ErrResiInvalid)

	t.Log("✅ Validasi resi tidak valid berhasil!")
}

// TestConcurrentUniqueEventIDs memverifikasi bahwa banyak event yang dibuat
// bersamaan semuanya memiliki UUID yang unik.
func TestConcurrentUniqueEventIDs(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	// Buat 5 resi dan proses masing-masing hingga scan-out
	for i := 1; i <= 5; i++ {
		resiID := fmt.Sprintf("RESI-BATCH-%03d", i)
		require.NoError(t, env.hubService.ScanIn(ctx, resiID, "HUB-JKT-01"))
		require.NoError(t, env.hubService.ScanOut(ctx, resiID, "HUB-JKT-01"))
	}

	// Total: 5 resi × 2 events = 10 tracking events
	// Semua ID harus unik
	allIDs := map[string]bool{}
	for i := 1; i <= 5; i++ {
		resiID := fmt.Sprintf("RESI-BATCH-%03d", i)
		events, err := env.repo.GetTrackingEvents(ctx, resiID)
		require.NoError(t, err)
		for _, e := range events {
			assert.False(t, allIDs[e.ID], "UUID duplikat ditemukan: %s", e.ID)
			allIDs[e.ID] = true
		}
	}
	assert.Len(t, allIDs, 10)

	t.Log("✅ UUID unik untuk semua event berhasil diverifikasi!")
}
