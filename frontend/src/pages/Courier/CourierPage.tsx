import React, { useState } from 'react';
import axios from 'axios';
import * as T from '../../types';

// =============================================================================
// Local Types
// =============================================================================
type DeliveryStatus = 'DELIVERED' | 'FAILED' | 'RETURNED';
type SidebarTab = 'tasks' | 'active' | 'completed' | 'earnings';

interface DeliveryTask {
  resi_id: string;
  recipient_name: string;
  destination: string;
  weight: number;
  deadline: string;
  service_type: 'EXPRESS' | 'REGULAR' | 'SAMEDAY' | 'NEXTDAY';
  item_type: string;
  is_urgent?: boolean;
  completed?: boolean;
}

// =============================================================================
// Mock Data — mirrors the API shape described in the contract
// =============================================================================
const MOCK_TASKS: DeliveryTask[] = [
  {
    resi_id: 'SLK-8923-J21',
    recipient_name: 'Ibu Siti Aminah',
    destination: 'Jl. Melati No. 12, Buah Batu\nBandung, 40287',
    weight: 0.5,
    deadline: 'Sebelum 12:00',
    service_type: 'EXPRESS',
    item_type: 'Dokumen Penting',
    is_urgent: true,
  },
  {
    resi_id: 'SLK-8924-M44',
    recipient_name: 'Bp. Agus Hermawan',
    destination: 'Komp. Permata Hijau Blok C4\nPasteur, Bandung',
    weight: 2.5,
    deadline: 'Sebelum 17:00',
    service_type: 'REGULAR',
    item_type: 'Pakaian',
  },
  {
    resi_id: 'SLK-8925-K90',
    recipient_name: 'Toko Elektronik Maju',
    destination: 'Jl. Jend. Sudirman No. 101\nAndir, Bandung',
    weight: 5.0,
    deadline: 'Sebelum 17:00',
    service_type: 'REGULAR',
    item_type: 'Elektronik (Fragile)',
  },
  {
    resi_id: 'RESI-982341029',
    recipient_name: 'Ibu Kartini Wulandari',
    destination: 'Jl. Diponegoro No. 55\nCicendo, Bandung',
    weight: 1.2,
    deadline: 'Sebelum 15:00',
    service_type: 'SAMEDAY',
    item_type: 'Makanan Kering',
    is_urgent: true,
  },
  {
    resi_id: 'RESI-982341030',
    recipient_name: 'Bp. Doni Prasetyo',
    destination: 'Jl. Raya Cimahi No. 88\nCimahi Utara',
    weight: 3.8,
    deadline: 'Sebelum 18:00',
    service_type: 'REGULAR',
    item_type: 'Spare Parts',
  },
];

const MOCK_PROOF_URL = 'https://bucket.s3.amazonaws.com/proofs/delivery-mock.jpg';

// =============================================================================
// Helpers
// =============================================================================
const getServiceBadgeStyle = (type: DeliveryTask['service_type']) => {
  switch (type) {
    case 'EXPRESS':
      return 'bg-emerald-50 text-emerald-600 border-emerald-200';
    case 'SAMEDAY':
      return 'bg-orange-50 text-orange-600 border-orange-200';
    case 'NEXTDAY':
      return 'bg-violet-50 text-violet-600 border-violet-200';
    default:
      return 'bg-sky-50 text-[#009ADA] border-sky-200';
  }
};

const getCardAccentColor = (type: DeliveryTask['service_type']) => {
  switch (type) {
    case 'EXPRESS': return 'border-l-emerald-400';
    case 'SAMEDAY': return 'border-l-orange-400';
    case 'NEXTDAY': return 'border-l-violet-400';
    default:        return 'border-l-[#2DB7F2]';
  }
};

function getInitials(name: string) {
  return name.split(' ').map(w => w[0]).join('').toUpperCase().substring(0, 2);
}

// =============================================================================
// Toast Banner
// =============================================================================
interface ToastProps { type: 'success' | 'error'; message: string; onClose: () => void; }
const ToastBanner: React.FC<ToastProps> = ({ type, message, onClose }) => (
  <div className={`fixed top-4 left-1/2 -translate-x-1/2 z-[100] flex items-center gap-3 px-5 py-3.5 rounded-2xl shadow-xl text-sm font-semibold animate-fadeIn max-w-sm w-full ${
    type === 'success' ? 'bg-emerald-500 text-white' : 'bg-rose-500 text-white'
  }`}>
    {type === 'success' ? (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-5 h-5 flex-shrink-0">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ) : (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-5 h-5 flex-shrink-0">
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
      </svg>
    )}
    <span className="flex-1">{message}</span>
    <button onClick={onClose} className="opacity-80 hover:opacity-100 transition-opacity">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
);

// =============================================================================
// Status Badge for bottom sheet options
// =============================================================================
const STATUS_OPTIONS: { value: DeliveryStatus; label: string; desc: string; color: string; activeColor: string }[] = [
  { value: 'DELIVERED', label: '✓ DELIVERED',  desc: 'Paket diterima oleh penerima', color: 'border-slate-200 text-slate-600 hover:border-emerald-300', activeColor: 'border-emerald-500 bg-emerald-50 text-emerald-700 ring-2 ring-emerald-200' },
  { value: 'FAILED',    label: '✕ GAGAL KIRIM', desc: 'Penerima tidak ada / alamat salah', color: 'border-slate-200 text-slate-600 hover:border-rose-300', activeColor: 'border-rose-500 bg-rose-50 text-rose-700 ring-2 ring-rose-200' },
  { value: 'RETURNED',  label: '↩ RETURNED',   desc: 'Paket dikembalikan ke HUB', color: 'border-slate-200 text-slate-600 hover:border-orange-300', activeColor: 'border-orange-500 bg-orange-50 text-orange-700 ring-2 ring-orange-200' },
];

// =============================================================================
// Sidebar nav items
// =============================================================================
const NAV_ITEMS: { id: SidebarTab; label: string; icon: React.ReactNode }[] = [
  {
    id: 'tasks',
    label: "Today's Tasks",
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25zM6.75 12h.008v.008H6.75V12zm0 3h.008v.008H6.75V15zm0 3h.008v.008H6.75V18z" />
      </svg>
    ),
  },
  {
    id: 'active',
    label: 'Active Deliveries',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.113c0-.732-.412-1.4-1.071-1.724a8.256 8.256 0 00-3.03-.69" />
      </svg>
    ),
  },
  {
    id: 'completed',
    label: 'Completed',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  {
    id: 'earnings',
    label: 'Earnings',
    icon: (
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
      </svg>
    ),
  },
];

// =============================================================================
// Main Component
// =============================================================================
const CourierPage: React.FC = () => {
  const [activeTab, setActiveTab]       = useState<SidebarTab>('tasks');
  const [sidebarOpen, setSidebarOpen]   = useState(false);        // mobile drawer
  const [searchQuery, setSearchQuery]   = useState('');
  const [tasks, setTasks]               = useState<DeliveryTask[]>(MOCK_TASKS);

  // Bottom sheet state
  const [sheetTask,   setSheetTask]     = useState<DeliveryTask | null>(null);
  const [pickedStatus, setPickedStatus] = useState<DeliveryStatus | null>(null);
  const [loading,     setLoading]       = useState(false);
  const [toast,       setToast]         = useState<{ type: 'success' | 'error'; msg: string } | null>(null);

  // Courier identity (would come from auth context in real app)
  const courierName = 'Budi';

  // Filtered tasks based on search
  const pendingTasks  = tasks.filter(t => !t.completed);
  const doneTasks     = tasks.filter(t =>  t.completed);
  const filteredTasks = pendingTasks.filter(t =>
    !searchQuery ||
    t.resi_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
    t.recipient_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    t.destination.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Show toast helper
  const showToast = (type: 'success' | 'error', msg: string) => {
    setToast({ type, msg });
    setTimeout(() => setToast(null), 4000);
  };

  // Open bottom sheet
  const openSheet = (task: DeliveryTask) => {
    setSheetTask(task);
    setPickedStatus(null);
  };

  // Close bottom sheet
  const closeSheet = () => {
    setSheetTask(null);
    setPickedStatus(null);
  };

  // Submit delivery status
  const handleSubmitStatus = async () => {
    if (!sheetTask || !pickedStatus) return;

    setLoading(true);
    const token = localStorage.getItem('token');
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const payload: T.DeliveryStatusRequest = {
      resi_id:   sheetTask.resi_id,
      status:    pickedStatus as T.TrackingStatus,
      proof_url: MOCK_PROOF_URL,
    };

    try {
      const res = await axios.post<T.SuccessResponse>(
        'http://localhost:8085/courier/delivery-status',
        payload,
        { headers }
      );
      showToast('success', res.data.message || 'Status berhasil diperbarui.');
      // Mark task as completed
      setTasks(prev =>
        prev.map(t => t.resi_id === sheetTask.resi_id ? { ...t, completed: true } : t)
      );
      closeSheet();
    } catch (err: any) {
      const msg =
        err.response?.data?.message ||
        err.response?.data?.error ||
        'Server tidak merespons. Pastikan backend berjalan di port 8085.';
      showToast('error', msg);
    } finally {
      setLoading(false);
    }
  };

  // ============================================================
  // Sidebar (desktop persistent, mobile drawer)
  // ============================================================
  const SidebarContent = () => (
    <aside className="flex flex-col h-full">
      {/* Logo */}
      <div className="p-5 border-b border-white/10">
        <div className="flex items-center gap-2.5">
          <div className="p-1.5 bg-white/20 rounded-lg">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="white" className="w-5 h-5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25" />
            </svg>
          </div>
          <div>
            <p className="text-white font-black text-sm leading-tight tracking-tight">SkyLogistics</p>
            <p className="text-white/60 text-[9px] font-bold uppercase tracking-widest">Courier Portal</p>
          </div>
        </div>
      </div>

      {/* Start New Route CTA */}
      <div className="px-4 pt-5">
        <button className="w-full flex items-center justify-center gap-2 py-3 bg-white text-[#009ADA] font-extrabold text-sm rounded-xl shadow-md shadow-sky-900/20 hover:bg-sky-50 active:scale-95 transition-all duration-150">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          Start New Route
        </button>
      </div>

      {/* Nav Items */}
      <nav className="flex-1 px-3 pt-6 space-y-1">
        {NAV_ITEMS.map((item) => {
          const isActive = activeTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => { setActiveTab(item.id); setSidebarOpen(false); }}
              className={`w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-semibold transition-all duration-150 ${
                isActive
                  ? 'bg-white/20 text-white shadow-sm'
                  : 'text-white/60 hover:text-white hover:bg-white/10'
              }`}
            >
              {item.icon}
              {item.label}
            </button>
          );
        })}
      </nav>

      {/* Settings */}
      <div className="px-3 pb-5">
        <button className="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-semibold text-white/60 hover:text-white hover:bg-white/10 transition-all duration-150">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
            <path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Settings
        </button>
      </div>
    </aside>
  );

  // ============================================================
  // Delivery Card
  // ============================================================
  const DeliveryCard: React.FC<{ task: DeliveryTask; isFirst?: boolean }> = ({ task, isFirst }) => (
    <div className={`relative bg-white border border-slate-100/60 border-l-4 ${getCardAccentColor(task.service_type)} rounded-2xl shadow-md shadow-sky-100/30 flex flex-col min-w-[280px] sm:min-w-[300px] max-w-xs flex-shrink-0 overflow-hidden`}>

      {/* Top: Service Badge + Deadline */}
      <div className="flex items-start justify-between px-5 pt-5 pb-3">
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-[10px] font-extrabold tracking-widest border uppercase ${getServiceBadgeStyle(task.service_type)}`}>
          {task.service_type}
        </span>
        <div className="text-right">
          <p className="text-[9px] font-bold text-slate-400 uppercase tracking-wider">Deadline</p>
          <p className={`text-xs font-extrabold flex items-center gap-1 justify-end ${task.is_urgent ? 'text-rose-500' : 'text-slate-600'}`}>
            {task.is_urgent && (
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-3 h-3">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            )}
            {task.deadline}
          </p>
        </div>
      </div>

      {/* Recipient + Resi */}
      <div className="px-5 pb-4 border-b border-slate-50">
        <h3 className="text-base font-black text-slate-900 leading-snug">{task.recipient_name}</h3>
        <p className="text-[11px] font-mono text-slate-400 mt-0.5">RESI: {task.resi_id}</p>
      </div>

      {/* Details */}
      <div className="px-5 py-4 space-y-3 flex-1">
        {/* Destination */}
        <div className="flex gap-2.5">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4 text-slate-400 flex-shrink-0 mt-0.5">
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
          </svg>
          <div>
            <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wide">Tujuan</p>
            <p className="text-xs font-semibold text-slate-700 leading-relaxed whitespace-pre-line">{task.destination}</p>
          </div>
        </div>

        {/* Package detail */}
        <div className="flex gap-2.5">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4 text-slate-400 flex-shrink-0 mt-0.5">
            <path strokeLinecap="round" strokeLinejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
          </svg>
          <div>
            <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wide">Detail Paket</p>
            <p className="text-xs font-semibold text-slate-700">{task.item_type} • {task.weight} Kg</p>
          </div>
        </div>
      </div>

      {/* Action Row */}
      <div className="flex items-center gap-3 px-5 pb-5">
        {/* Phone */}
        <button className="p-2.5 rounded-xl border border-slate-200 text-slate-400 hover:text-[#2DB7F2] hover:border-sky-200 hover:bg-sky-50 transition-all duration-150">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4">
            <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 6.75c0 8.284 6.716 15 15 15h2.25a2.25 2.25 0 002.25-2.25v-1.372c0-.516-.351-.966-.852-1.091l-4.423-1.106c-.44-.11-.902.055-1.173.417l-.97 1.293c-.282.376-.769.542-1.21.38a12.035 12.035 0 01-7.143-7.143c-.162-.441.004-.928.38-1.21l1.293-.97c.363-.271.527-.734.417-1.173L6.963 3.102a1.125 1.125 0 00-1.091-.852H4.5A2.25 2.25 0 002.25 4.5v2.25z" />
          </svg>
        </button>

        {/* Update Status */}
        <button
          id={`btn-update-status-${task.resi_id}`}
          onClick={() => openSheet(task)}
          className={`flex-1 py-2.5 rounded-xl text-sm font-bold transition-all duration-150 active:scale-95 ${
            isFirst
              ? 'bg-gradient-to-r from-[#2DB7F2] to-[#009ADA] text-white shadow-md shadow-sky-200/50 hover:from-[#009ADA] hover:to-[#007BB5]'
              : 'border border-[#2DB7F2] text-[#009ADA] hover:bg-sky-50'
          }`}
        >
          Update Status
        </button>
      </div>
    </div>
  );

  // ============================================================
  // Content Panels
  // ============================================================
  const TodaysTasks = () => (
    <div className="space-y-6">
      {/* Page heading */}
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight">
            Daftar Tugas Pengantaran
          </h1>
          <div className="flex items-center gap-2 mt-1.5 flex-wrap">
            <span className="text-slate-500 text-sm">👋 Halo, Kurir {courierName} —</span>
            <span className="text-sm font-bold text-[#009ADA] bg-sky-50 px-2.5 py-0.5 rounded-full border border-sky-100">
              {pendingTasks.length} Sisa Tugas Hari Ini
            </span>
          </div>
        </div>

        {/* Filter + Sort */}
        <div className="flex items-center gap-2 flex-shrink-0">
          <button className="flex items-center gap-1.5 px-4 py-2 border border-slate-200 rounded-xl text-sm font-semibold text-slate-600 hover:bg-slate-50 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 3c2.755 0 5.455.232 8.083.678.533.09.917.556.917 1.096v1.044a2.25 2.25 0 01-.659 1.591l-5.432 5.432a2.25 2.25 0 00-.659 1.591v2.927a2.25 2.25 0 01-1.244 2.013L9.75 21v-6.568a2.25 2.25 0 00-.659-1.591L3.659 7.409A2.25 2.25 0 013 5.818V4.774c0-.54.384-1.006.917-1.096A48.32 48.32 0 0112 3z" />
            </svg>
            Filter
          </button>
          <button className="flex items-center gap-1.5 px-4 py-2 border border-slate-200 rounded-xl text-sm font-semibold text-slate-600 hover:bg-slate-50 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4">
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 7.5L7.5 3m0 0L12 7.5M7.5 3v13.5m13.5 0L16.5 21m0 0L12 16.5m4.5 4.5V7.5" />
            </svg>
            Urutkan
          </button>
        </div>
      </div>

      {/* Task Cards – horizontal scroll on mobile/tablet, grid on large screens */}
      {filteredTasks.length > 0 ? (
        <>
          {/* Mobile/Tablet: horizontal scroll */}
          <div className="flex gap-4 overflow-x-auto pb-3 snap-x snap-mandatory lg:hidden scrollbar-hide">
            {filteredTasks.map((task, i) => (
              <div key={task.resi_id} className="snap-start flex-shrink-0">
                <DeliveryCard task={task} isFirst={i === 0} />
              </div>
            ))}
          </div>

          {/* Desktop: responsive grid */}
          <div className="hidden lg:grid grid-cols-2 xl:grid-cols-3 gap-5">
            {filteredTasks.map((task, i) => (
              <DeliveryCard key={task.resi_id} task={task} isFirst={i === 0} />
            ))}
          </div>
        </>
      ) : (
        <div className="flex flex-col items-center justify-center py-20 text-slate-400">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-14 h-14 mb-4 text-slate-300">
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-lg font-bold">Semua tugas selesai!</p>
          <p className="text-sm mt-1">Tidak ada paket yang menunggu pengantaran.</p>
        </div>
      )}

      {/* Completed section */}
      {doneTasks.length > 0 && (
        <div className="mt-4">
          <p className="text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">
            ✓ Selesai Hari Ini ({doneTasks.length})
          </p>
          <div className="space-y-2">
            {doneTasks.map(task => (
              <div key={task.resi_id} className="flex items-center justify-between px-5 py-3 bg-white border border-slate-100 rounded-xl text-sm opacity-60">
                <div>
                  <span className="font-mono font-bold text-slate-500">{task.resi_id}</span>
                  <span className="mx-2 text-slate-300">·</span>
                  <span className="text-slate-500">{task.recipient_name}</span>
                </div>
                <span className="text-emerald-500 font-bold text-xs">SELESAI</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  const PlaceholderPanel: React.FC<{ title: string; desc: string }> = ({ title, desc }) => (
    <div className="flex flex-col items-center justify-center h-64 text-slate-400 space-y-2">
      <p className="text-xl font-black text-slate-600">{title}</p>
      <p className="text-sm">{desc}</p>
    </div>
  );

  // ============================================================
  // Render
  // ============================================================
  return (
    <div className="flex h-screen bg-slate-50 font-sans overflow-hidden">

      {/* ══════════════════════════════
          DESKTOP SIDEBAR
      ══════════════════════════════ */}
      <div className="hidden md:flex flex-col w-52 flex-shrink-0 bg-gradient-to-b from-[#2DB7F2] to-[#007BB5] shadow-xl shadow-sky-900/20">
        <SidebarContent />
      </div>

      {/* ══════════════════════════════
          MOBILE DRAWER OVERLAY
      ══════════════════════════════ */}
      {sidebarOpen && (
        <div className="fixed inset-0 z-50 flex md:hidden">
          <div
            className="absolute inset-0 bg-black/40 backdrop-blur-sm"
            onClick={() => setSidebarOpen(false)}
          />
          <div className="relative w-56 bg-gradient-to-b from-[#2DB7F2] to-[#007BB5] shadow-2xl animate-slideInLeft">
            <SidebarContent />
          </div>
        </div>
      )}

      {/* ══════════════════════════════
          MAIN AREA
      ══════════════════════════════ */}
      <div className="flex flex-col flex-1 min-w-0 overflow-hidden">

        {/* ── Topbar ── */}
        <header className="flex items-center gap-4 px-4 sm:px-6 py-4 bg-white border-b border-slate-100 shadow-sm shadow-slate-100/40 flex-shrink-0">

          {/* Hamburger (mobile only) */}
          <button
            className="md:hidden p-2 text-slate-500 hover:text-slate-700 hover:bg-slate-50 rounded-xl transition-colors"
            onClick={() => setSidebarOpen(true)}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
            </svg>
          </button>

          {/* Search Bar */}
          <div className="relative flex-1 max-w-lg mx-auto">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400">
              <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
            </svg>
            <input
              type="text"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              placeholder="Cari nomor resi, nama, atau alamat…"
              className="w-full pl-10 pr-4 py-2.5 border border-slate-200 rounded-xl text-sm text-slate-700 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-[#2DB7F2]/40 focus:border-[#2DB7F2] transition-all"
            />
          </div>

          {/* Icons */}
          <div className="flex items-center gap-2 flex-shrink-0">
            <button className="relative p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded-xl transition-colors">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
              </svg>
              <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 bg-rose-500 rounded-full" />
            </button>
            <button className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded-xl transition-colors hidden sm:block">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
              </svg>
            </button>
            <div className="w-9 h-9 rounded-full bg-gradient-to-br from-[#2DB7F2] to-[#009ADA] flex items-center justify-center text-white text-xs font-black shadow-sm border-2 border-white">
              {getInitials('Kurir ' + courierName)}
            </div>
          </div>
        </header>

        {/* ── Scrollable Page Content ── */}
        <main className="flex-1 overflow-y-auto px-4 sm:px-6 lg:px-8 py-6 pb-20 md:pb-6">
          {activeTab === 'tasks'     && <TodaysTasks />}
          {activeTab === 'active'    && <PlaceholderPanel title="Active Deliveries" desc="Pengiriman yang sedang dalam perjalanan." />}
          {activeTab === 'completed' && <PlaceholderPanel title="Completed" desc="Daftar tugas yang sudah selesai hari ini." />}
          {activeTab === 'earnings'  && <PlaceholderPanel title="Earnings" desc="Ringkasan pendapatan harian dan mingguan." />}
        </main>

        {/* ── Footer ── */}
        <footer className="hidden md:flex items-center justify-between px-6 py-3.5 bg-white border-t border-slate-100 text-xs text-slate-400 font-medium flex-shrink-0">
          <span>© 2024 SkyLogistics Solutions. All rights reserved.</span>
          <div className="flex items-center gap-5">
            <a href="#" className="hover:text-slate-600 transition-colors">Privacy Policy</a>
            <a href="#" className="hover:text-slate-600 transition-colors">Terms of Service</a>
            <a href="#" className="hover:text-slate-600 transition-colors">Support</a>
          </div>
        </footer>
      </div>

      {/* ══════════════════════════════
          MOBILE BOTTOM NAV BAR
      ══════════════════════════════ */}
      <nav className="fixed bottom-0 inset-x-0 md:hidden bg-white border-t border-slate-100 flex items-center justify-around py-2 z-40 shadow-xl shadow-slate-200/50">
        {NAV_ITEMS.map(item => {
          const isActive = activeTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => setActiveTab(item.id)}
              className={`flex flex-col items-center gap-1 px-3 py-1.5 rounded-xl transition-all duration-150 ${
                isActive ? 'text-[#2DB7F2]' : 'text-slate-400 hover:text-slate-600'
              }`}
            >
              {item.icon}
              <span className="text-[9px] font-bold">{item.label.split(' ')[0]}</span>
            </button>
          );
        })}
      </nav>

      {/* ══════════════════════════════
          BOTTOM SHEET: Update Status
      ══════════════════════════════ */}
      {sheetTask && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 animate-fadeIn"
            onClick={closeSheet}
          />

          {/* Sheet */}
          <div className="fixed bottom-0 left-0 right-0 md:inset-0 md:flex md:items-center md:justify-center z-50">
            <div className="bg-white md:rounded-2xl md:max-w-md md:w-full w-full rounded-t-3xl shadow-2xl animate-slideUp md:animate-fadeIn max-h-[90vh] overflow-y-auto">

              {/* Handle bar (mobile) */}
              <div className="flex justify-center pt-3 pb-1 md:hidden">
                <div className="w-10 h-1 bg-slate-200 rounded-full" />
              </div>

              {/* Sheet Header */}
              <div className="flex items-start justify-between px-6 pt-5 pb-4 border-b border-slate-100">
                <div>
                  <h2 className="text-lg font-black text-slate-900">Update Status Kirim</h2>
                  <p className="text-xs font-mono text-slate-400 mt-0.5">{sheetTask.resi_id}</p>
                  <p className="text-sm font-bold text-slate-600 mt-1">{sheetTask.recipient_name}</p>
                </div>
                <button onClick={closeSheet} className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded-xl transition-colors">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-5 h-5">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              {/* Sheet Body */}
              <div className="px-6 py-5 space-y-5">

                {/* Status Selection */}
                <div>
                  <p className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-3">Pilih Status Pengiriman</p>
                  <div className="space-y-2.5">
                    {STATUS_OPTIONS.map(opt => (
                      <button
                        key={opt.value}
                        id={`status-btn-${opt.value.toLowerCase()}`}
                        onClick={() => setPickedStatus(opt.value)}
                        className={`w-full flex items-start gap-3 p-4 rounded-xl border-2 text-left transition-all duration-150 ${
                          pickedStatus === opt.value ? opt.activeColor : opt.color
                        }`}
                      >
                        <div className={`w-4 h-4 rounded-full border-2 mt-0.5 flex-shrink-0 flex items-center justify-center ${
                          pickedStatus === opt.value ? 'border-current' : 'border-slate-300'
                        }`}>
                          {pickedStatus === opt.value && (
                            <div className="w-2 h-2 rounded-full bg-current" />
                          )}
                        </div>
                        <div>
                          <p className="text-sm font-extrabold">{opt.label}</p>
                          <p className="text-xs font-medium opacity-70 mt-0.5">{opt.desc}</p>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Mock Proof URL */}
                <div>
                  <p className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-2">Bukti Foto (URL)</p>
                  <div className="flex items-center gap-3 p-3.5 bg-slate-50 border border-slate-200 rounded-xl">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5 text-slate-400 flex-shrink-0">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5zm10.5-11.25h.008v.008h-.008V8.25zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" />
                    </svg>
                    <span className="text-xs font-medium text-slate-500 truncate">{MOCK_PROOF_URL}</span>
                    <span className="text-[10px] font-bold text-emerald-500 bg-emerald-50 border border-emerald-100 px-2 py-0.5 rounded-full flex-shrink-0">MOCK</span>
                  </div>
                </div>

                {/* Submit Button */}
                <button
                  id="btn-simpan-status"
                  onClick={handleSubmitStatus}
                  disabled={!pickedStatus || loading}
                  className="w-full py-3.5 bg-gradient-to-r from-[#2DB7F2] to-[#009ADA] hover:from-[#009ADA] hover:to-[#007BB5] text-white font-extrabold rounded-xl shadow-md shadow-sky-200/50 active:scale-[0.98] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {loading ? (
                    <>
                      <svg className="animate-spin w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                      </svg>
                      Menyimpan…
                    </>
                  ) : (
                    'Simpan Status'
                  )}
                </button>

                {/* Cancel */}
                <button
                  onClick={closeSheet}
                  className="w-full py-3 text-sm font-bold text-slate-400 hover:text-slate-600 transition-colors"
                >
                  Batal
                </button>
              </div>
            </div>
          </div>
        </>
      )}

      {/* ══════════════════════════════
          GLOBAL TOAST
      ══════════════════════════════ */}
      {toast && (
        <ToastBanner
          type={toast.type}
          message={toast.msg}
          onClose={() => setToast(null)}
        />
      )}
    </div>
  );
};

export default CourierPage;
