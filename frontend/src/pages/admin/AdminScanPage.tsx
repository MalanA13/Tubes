import React, { useState } from 'react';
import axios from 'axios';
import * as T from '../../types';

// =============================================================================
// Types for Hub Scan Log (local mock, mirrors future API shape)
// =============================================================================
type ScanAction = 'SCAN-IN' | 'SCAN-OUT';

interface ScanLogEntry {
  id: string;
  resi_id: string;
  timestamp: string;
  action: ScanAction;
  operator_name: string;
}

// =============================================================================
// Hub Location Options (mirrors backend hub_id values)
// =============================================================================
const HUB_OPTIONS: { label: string; value: string }[] = [
  { label: 'Jakarta Timur – HUB A1', value: 'HUB-JKT-01' },
  { label: 'Bandung Kota – HUB B1',  value: 'HUB-BDO-01' },
  { label: 'Surabaya Utara – HUB C1', value: 'HUB-SUB-01' },
  { label: 'Medan Barat – HUB D1',   value: 'HUB-KNO-01' },
  { label: 'Makassar – HUB E1',      value: 'HUB-UPG-01' },
];

// =============================================================================
// Mock scan history (will be replaced by real API when available)
// =============================================================================
const INITIAL_SCAN_LOGS: ScanLogEntry[] = [
  { id: '1', resi_id: 'SLX-982341029', timestamp: '24 Oct 2024, 14:22:10', action: 'SCAN-IN',  operator_name: 'Bambang Susanto' },
  { id: '2', resi_id: 'SLX-982341030', timestamp: '24 Oct 2024, 14:15:45', action: 'SCAN-OUT', operator_name: 'Anita Rahayu' },
  { id: '3', resi_id: 'SLX-982341031', timestamp: '24 Oct 2024, 14:08:12', action: 'SCAN-IN',  operator_name: 'Bambang Susanto' },
  { id: '4', resi_id: 'SLX-982341032', timestamp: '24 Oct 2024, 13:55:30', action: 'SCAN-OUT', operator_name: 'Dedi Pratama' },
];

// =============================================================================
// Helper – format timestamp from ISO
// =============================================================================
function formatTimestamp(iso: string): string {
  const d = new Date(iso);
  const day   = d.getDate().toString().padStart(2, '0');
  const month = d.toLocaleString('en-US', { month: 'short' });
  const year  = d.getFullYear();
  const hh    = d.getHours().toString().padStart(2, '0');
  const mm    = d.getMinutes().toString().padStart(2, '0');
  const ss    = d.getSeconds().toString().padStart(2, '0');
  return `${day} ${month} ${year}, ${hh}:${mm}:${ss}`;
}

// =============================================================================
// Helper – operator initials avatar
// =============================================================================
function getInitials(name: string): string {
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .substring(0, 2);
}

// =============================================================================
// Sub-component: Result Toast Banner
// =============================================================================
interface ToastProps {
  type: 'success' | 'error';
  message: string;
  onClose: () => void;
}
const ToastBanner: React.FC<ToastProps> = ({ type, message, onClose }) => (
  <div
    className={`flex items-start gap-3 p-4 rounded-xl border text-sm font-semibold animate-fadeIn ${
      type === 'success'
        ? 'bg-emerald-50 border-emerald-100 text-emerald-700'
        : 'bg-rose-50 border-rose-100 text-rose-700'
    }`}
  >
    {type === 'success' ? (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-5 h-5 flex-shrink-0 mt-0.5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ) : (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-5 h-5 flex-shrink-0 mt-0.5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
      </svg>
    )}
    <span className="flex-1">{message}</span>
    <button onClick={onClose} className="hover:opacity-70 transition-opacity">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
);

// =============================================================================
// Main Component
// =============================================================================
const AdminScanPage: React.FC = () => {
  // HUB selector (shared for both panels)
  const [selectedHub, setSelectedHub] = useState(HUB_OPTIONS[0].value);
  const [hubDropdownOpen, setHubDropdownOpen] = useState(false);

  // Scan-In state
  const [resiIn,        setResiIn]        = useState('');
  const [loadingIn,     setLoadingIn]     = useState(false);
  const [toastIn,       setToastIn]       = useState<{ type: 'success' | 'error'; msg: string } | null>(null);

  // Scan-Out state
  const [resiOut,       setResiOut]       = useState('');
  const [loadingOut,    setLoadingOut]    = useState(false);
  const [toastOut,      setToastOut]      = useState<{ type: 'success' | 'error'; msg: string } | null>(null);

  // Scan log (prepended after each successful scan)
  const [scanLogs, setScanLogs] = useState<ScanLogEntry[]>(INITIAL_SCAN_LOGS);

  // --------------------------------------------------------
  // Shared API call factory
  // --------------------------------------------------------
  const buildHeaders = (): Record<string, string> => {
    const token = localStorage.getItem('token');
    const h: Record<string, string> = { 'Content-Type': 'application/json' };
    if (token) h['Authorization'] = `Bearer ${token}`;
    return h;
  };

  const appendLog = (resi_id: string, action: ScanAction) => {
    const newEntry: ScanLogEntry = {
      id:            Date.now().toString(),
      resi_id,
      timestamp:     formatTimestamp(new Date().toISOString()),
      action,
      operator_name: 'Admin User',
    };
    setScanLogs((prev) => [newEntry, ...prev]);
  };

  // --------------------------------------------------------
  // Scan-In handler
  // --------------------------------------------------------
  const handleScanIn = async () => {
    if (!resiIn.trim()) {
      setToastIn({ type: 'error', msg: 'Nomor resi tidak boleh kosong.' });
      return;
    }

    setLoadingIn(true);
    setToastIn(null);

    const payload: T.ScanRequest = { resi_id: resiIn.trim(), hub_id: selectedHub };

    try {
      const res = await axios.post<T.SuccessResponse>(
        'http://localhost:8084/hub/scan-in',
        payload,
        { headers: buildHeaders() }
      );
      setToastIn({ type: 'success', msg: res.data.message || 'Scan-in berhasil dicatat.' });
      appendLog(resiIn.trim(), 'SCAN-IN');
      setResiIn('');
    } catch (err: any) {
      const msg =
        err.response?.data?.message ||
        err.response?.data?.error ||
        'Server tidak merespons. Pastikan backend berjalan di port 8084.';
      setToastIn({ type: 'error', msg });
    } finally {
      setLoadingIn(false);
    }
  };

  // --------------------------------------------------------
  // Scan-Out handler
  // --------------------------------------------------------
  const handleScanOut = async () => {
    if (!resiOut.trim()) {
      setToastOut({ type: 'error', msg: 'Nomor resi tidak boleh kosong.' });
      return;
    }

    setLoadingOut(true);
    setToastOut(null);

    const payload: T.ScanRequest = { resi_id: resiOut.trim(), hub_id: selectedHub };

    try {
      const res = await axios.post<T.SuccessResponse>(
        'http://localhost:8084/hub/scan-out',
        payload,
        { headers: buildHeaders() }
      );
      setToastOut({ type: 'success', msg: res.data.message || 'Scan-out berhasil dicatat.' });
      appendLog(resiOut.trim(), 'SCAN-OUT');
      setResiOut('');
    } catch (err: any) {
      const msg =
        err.response?.data?.message ||
        err.response?.data?.error ||
        'Server tidak merespons. Pastikan backend berjalan di port 8084.';
      setToastOut({ type: 'error', msg });
    } finally {
      setLoadingOut(false);
    }
  };

  // --------------------------------------------------------
  // Keyboard: submit on Enter
  // --------------------------------------------------------
  const onKeyIn  = (e: React.KeyboardEvent) => { if (e.key === 'Enter') handleScanIn(); };
  const onKeyOut = (e: React.KeyboardEvent) => { if (e.key === 'Enter') handleScanOut(); };

  // Current hub label
  const currentHubLabel = HUB_OPTIONS.find((h) => h.value === selectedHub)?.label ?? selectedHub;

  // --------------------------------------------------------
  // Render
  // --------------------------------------------------------
  return (
    <div className="min-h-screen bg-slate-50/50 py-8 px-4 sm:px-6 lg:px-8 animate-fadeIn">
      <div className="max-w-6xl mx-auto space-y-8">

        {/* ══════════════════════════════════════════════════════════
            HEADER ROW: Title (left) + HUB Dropdown (right)
        ══════════════════════════════════════════════════════════ */}
        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-6">

          {/* Page Title */}
          <div>
            <h1 className="text-3xl font-black text-slate-900 tracking-tight">
              Manajemen Scanning HUB
            </h1>
            <p className="text-slate-500 mt-1 text-sm">
              Kelola alur masuk dan keluar paket pada titik distribusi.
            </p>
          </div>

          {/* HUB Location Dropdown */}
          <div className="relative flex-shrink-0">
            <p className="text-xs font-bold text-slate-500 mb-1.5">Pilih Lokasi Gudang/HUB Saat Ini</p>
            <button
              id="hub-selector"
              onClick={() => setHubDropdownOpen((v) => !v)}
              className="flex items-center justify-between gap-4 min-w-[240px] px-4 py-3 bg-white border border-slate-200 rounded-xl shadow-sm hover:shadow-md hover:border-sky-200 transition-all duration-150 text-sm font-semibold text-slate-700"
            >
              <span>{currentHubLabel}</span>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={2.5}
                stroke="currentColor"
                className={`w-4 h-4 text-slate-400 transition-transform duration-200 ${hubDropdownOpen ? 'rotate-180' : ''}`}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
              </svg>
            </button>

            {hubDropdownOpen && (
              <ul className="absolute right-0 mt-1 w-full bg-white border border-slate-100 rounded-xl shadow-xl shadow-slate-200/50 z-50 overflow-hidden animate-fadeIn">
                {HUB_OPTIONS.map((hub) => (
                  <li key={hub.value}>
                    <button
                      onClick={() => {
                        setSelectedHub(hub.value);
                        setHubDropdownOpen(false);
                      }}
                      className={`w-full text-left px-4 py-3 text-sm font-semibold transition-colors ${
                        selectedHub === hub.value
                          ? 'bg-sky-50 text-[#009ADA]'
                          : 'text-slate-700 hover:bg-slate-50'
                      }`}
                    >
                      {hub.label}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* ══════════════════════════════════════════════════════════
            SCAN PANELS: Scan-In (left) | Scan-Out (right)
        ══════════════════════════════════════════════════════════ */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

          {/* ── SCAN-IN PANEL ── */}
          <div className="bg-white rounded-2xl shadow-md shadow-sky-100/40 overflow-hidden border border-slate-100/60 flex flex-col">
            {/* Green top accent bar */}
            <div className="h-1 bg-gradient-to-r from-emerald-400 to-teal-400 w-full" />

            <div className="p-6 flex flex-col gap-5 flex-1">
              {/* Panel Title */}
              <div className="flex items-center gap-3">
                <div className="p-2 bg-emerald-50 text-emerald-500 rounded-xl border border-emerald-100">
                  {/* Arrow-right-into-box icon */}
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.2} stroke="currentColor" className="w-5 h-5">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9" />
                  </svg>
                </div>
                <h2 className="text-lg font-bold text-slate-800">SCAN MASUK (Scan-In)</h2>
              </div>

              {/* Toast */}
              {toastIn && (
                <ToastBanner
                  type={toastIn.type}
                  message={toastIn.msg}
                  onClose={() => setToastIn(null)}
                />
              )}

              {/* Resi Input */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-wide">
                  Nomor Resi / AWB
                </label>
                <div className="relative">
                  <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-[#2DB7F2]">
                    {/* Barcode icon */}
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 013.75 9.375v-4.5zM3.75 14.625c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5a1.125 1.125 0 01-1.125-1.125v-4.5zM13.5 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0113.5 9.375v-4.5z" />
                      <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 6.75h.75v.75h-.75v-.75zM6.75 16.5h.75v.75h-.75v-.75zM16.5 6.75h.75v.75h-.75v-.75zM13.5 13.5h.75v.75h-.75v-.75zM13.5 19.5h.75v.75h-.75v-.75zM19.5 13.5h.75v.75h-.75v-.75zM19.5 19.5h.75v.75h-.75v-.75zM16.5 16.5h.75v.75h-.75v-.75z" />
                    </svg>
                  </span>
                  <input
                    id="resi-in-input"
                    type="text"
                    value={resiIn}
                    onChange={(e) => setResiIn(e.target.value)}
                    onKeyDown={onKeyIn}
                    placeholder="Scan atau ketik ID resi…"
                    className="w-full pl-11 pr-4 py-3 border border-slate-200 rounded-xl text-sm font-semibold text-slate-700 placeholder-[#2DB7F2]/60 focus:outline-none focus:ring-2 focus:ring-[#2DB7F2]/40 focus:border-[#2DB7F2] transition-all duration-150"
                  />
                </div>
              </div>

              {/* Info hint */}
              <div className="flex items-start gap-3 p-4 bg-slate-50 border border-slate-100 rounded-xl text-xs text-slate-500 font-medium">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4 text-slate-400 flex-shrink-0 mt-0.5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
                </svg>
                Pastikan paket diterima dalam kondisi fisik yang baik sebelum melakukan scan.
              </div>

              {/* Submit button */}
              <button
                id="btn-scan-in"
                onClick={handleScanIn}
                disabled={loadingIn}
                className="mt-auto flex items-center justify-center gap-2.5 w-full py-3.5 bg-gradient-to-r from-[#2DB7F2] to-[#009ADA] hover:from-[#009ADA] hover:to-[#007BB5] text-white font-bold rounded-xl shadow-md shadow-sky-200/50 active:scale-[0.98] transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
              >
                {loadingIn ? (
                  <>
                    <svg className="animate-spin w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    <span>Menyimpan…</span>
                  </>
                ) : (
                  <>
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 3.75H6.912a2.25 2.25 0 00-2.15 1.588L2.35 13.177a2.25 2.25 0 00-.1.661V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18v-4.162c0-.224-.034-.447-.1-.661L19.24 5.338a2.25 2.25 0 00-2.15-1.588H15M2.25 13.5h3.86a2.25 2.25 0 012.012 1.244l.256.512a2.25 2.25 0 002.013 1.244h3.218a2.25 2.25 0 002.013-1.244l.256-.512a2.25 2.25 0 012.013-1.244h3.859M12 3v8.25m0 0l-3-3m3 3l3-3" />
                    </svg>
                    Simpan Scan Masuk
                  </>
                )}
              </button>
            </div>
          </div>

          {/* ── SCAN-OUT PANEL ── */}
          <div className="bg-white rounded-2xl shadow-md shadow-sky-100/40 overflow-hidden border border-slate-100/60 flex flex-col">
            {/* Purple top accent bar */}
            <div className="h-1 bg-gradient-to-r from-violet-400 to-purple-500 w-full" />

            <div className="p-6 flex flex-col gap-5 flex-1">
              {/* Panel Title */}
              <div className="flex items-center gap-3">
                <div className="p-2 bg-violet-50 text-violet-500 rounded-xl border border-violet-100">
                  {/* Arrow-left-from-box icon */}
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.2} stroke="currentColor" className="w-5 h-5">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75" />
                  </svg>
                </div>
                <h2 className="text-lg font-bold text-slate-800">SCAN KELUAR (Scan-Out)</h2>
              </div>

              {/* Toast */}
              {toastOut && (
                <ToastBanner
                  type={toastOut.type}
                  message={toastOut.msg}
                  onClose={() => setToastOut(null)}
                />
              )}

              {/* Resi Input */}
              <div className="space-y-1.5">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-wide">
                  Nomor Resi / AWB
                </label>
                <div className="relative">
                  <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-[#2DB7F2]">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 013.75 9.375v-4.5zM3.75 14.625c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5a1.125 1.125 0 01-1.125-1.125v-4.5zM13.5 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0113.5 9.375v-4.5z" />
                      <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 6.75h.75v.75h-.75v-.75zM6.75 16.5h.75v.75h-.75v-.75zM16.5 6.75h.75v.75h-.75v-.75zM13.5 13.5h.75v.75h-.75v-.75zM13.5 19.5h.75v.75h-.75v-.75zM19.5 13.5h.75v.75h-.75v-.75zM19.5 19.5h.75v.75h-.75v-.75zM16.5 16.5h.75v.75h-.75v-.75z" />
                    </svg>
                  </span>
                  <input
                    id="resi-out-input"
                    type="text"
                    value={resiOut}
                    onChange={(e) => setResiOut(e.target.value)}
                    onKeyDown={onKeyOut}
                    placeholder="Scan atau ketik ID resi…"
                    className="w-full pl-11 pr-4 py-3 border border-slate-200 rounded-xl text-sm font-semibold text-slate-700 placeholder-[#2DB7F2]/60 focus:outline-none focus:ring-2 focus:ring-[#2DB7F2]/40 focus:border-[#2DB7F2] transition-all duration-150"
                  />
                </div>
              </div>

              {/* Info hint */}
              <div className="flex items-start gap-3 p-4 bg-slate-50 border border-slate-100 rounded-xl text-xs text-slate-500 font-medium">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4 text-slate-400 flex-shrink-0 mt-0.5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.113c0-.732-.412-1.4-1.071-1.724a8.256 8.256 0 00-3.03-.69" />
                </svg>
                Input ID Kurir atau Fleet No. untuk menghubungkan pengiriman ke rute selanjutnya.
              </div>

              {/* Submit button */}
              <button
                id="btn-scan-out"
                onClick={handleScanOut}
                disabled={loadingOut}
                className="mt-auto flex items-center justify-center gap-2.5 w-full py-3.5 bg-gradient-to-r from-[#2DB7F2] to-[#009ADA] hover:from-[#009ADA] hover:to-[#007BB5] text-white font-bold rounded-xl shadow-md shadow-sky-200/50 active:scale-[0.98] transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
              >
                {loadingOut ? (
                  <>
                    <svg className="animate-spin w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    <span>Menyimpan…</span>
                  </>
                ) : (
                  <>
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 3.75H6.912a2.25 2.25 0 00-2.15 1.588L2.35 13.177a2.25 2.25 0 00-.1.661V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18v-4.162c0-.224-.034-.447-.1-.661L19.24 5.338a2.25 2.25 0 00-2.15-1.588H15M2.25 13.5h3.86a2.25 2.25 0 012.012 1.244l.256.512a2.25 2.25 0 002.013 1.244h3.218a2.25 2.25 0 002.013-1.244l.256-.512a2.25 2.25 0 012.013-1.244h3.859M12 3v8.25m0 0l-3-3m3 3l3-3" />
                    </svg>
                    Simpan Scan Keluar
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* ══════════════════════════════════════════════════════════
            SCAN HISTORY TABLE
        ══════════════════════════════════════════════════════════ */}
        <div className="bg-white rounded-2xl shadow-md shadow-sky-100/40 border border-slate-100/60 overflow-hidden">

          {/* Table header row */}
          <div className="flex items-center justify-between px-6 py-5 border-b border-slate-100">
            <div className="flex items-center gap-2.5 text-slate-800 font-bold">
              {/* Clock icon */}
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5 text-[#2DB7F2]">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Riwayat Scan Terakhir di Hub Ini</span>
            </div>
            <button className="flex items-center gap-1.5 text-xs font-bold text-[#009ADA] hover:text-[#2DB7F2] transition-colors">
              Lihat Semua Riwayat
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-3.5 h-3.5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
              </svg>
            </button>
          </div>

          {/* Table */}
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-slate-50/70 text-slate-400 text-xs font-bold tracking-widest uppercase border-b border-slate-100">
                  <th className="px-6 py-3.5">Resi ID</th>
                  <th className="px-6 py-3.5">Timestamp</th>
                  <th className="px-6 py-3.5">Action</th>
                  <th className="px-6 py-3.5">Operator Name</th>
                  <th className="px-6 py-3.5 text-right">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50">
                {scanLogs.map((log) => {
                  const initials = getInitials(log.operator_name);
                  const isScanIn = log.action === 'SCAN-IN';

                  return (
                    <tr key={log.id} className="hover:bg-slate-50/50 transition-colors">
                      {/* Resi ID */}
                      <td className="px-6 py-4">
                        <span className="font-bold text-[#009ADA] font-mono text-sm tracking-wide">
                          {log.resi_id}
                        </span>
                      </td>

                      {/* Timestamp */}
                      <td className="px-6 py-4 text-sm font-medium text-slate-500">
                        {log.timestamp}
                      </td>

                      {/* Action Badge */}
                      <td className="px-6 py-4">
                        <span
                          className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold border ${
                            isScanIn
                              ? 'bg-emerald-50 border-emerald-100 text-emerald-600'
                              : 'bg-violet-50 border-violet-100 text-violet-600'
                          }`}
                        >
                          <span
                            className={`w-1.5 h-1.5 rounded-full ${
                              isScanIn ? 'bg-emerald-500' : 'bg-violet-500'
                            }`}
                          />
                          {log.action}
                        </span>
                      </td>

                      {/* Operator */}
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2.5">
                          <div className="w-7 h-7 rounded-full bg-sky-100 text-[#009ADA] flex items-center justify-center text-[10px] font-black flex-shrink-0 border border-sky-200">
                            {initials}
                          </div>
                          <span className="text-sm font-semibold text-slate-700">
                            {log.operator_name}
                          </span>
                        </div>
                      </td>

                      {/* Status checkmark */}
                      <td className="px-6 py-4 text-right">
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 24 24"
                          strokeWidth={2}
                          stroke="currentColor"
                          className="w-6 h-6 text-emerald-500 inline-block"
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      </td>
                    </tr>
                  );
                })}

                {scanLogs.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center text-slate-400 text-sm font-medium">
                      Belum ada riwayat scan di sesi ini.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

      </div>
    </div>
  );
};

export default AdminScanPage;
