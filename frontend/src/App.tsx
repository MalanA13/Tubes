import React, { useState } from 'react';
import TrackingPage from './pages/customer/TrackingPage';
import CreateOrderPage from './pages/customer/CreateOrderPage';
import RegisterPage from './pages/auth/RegisterPage';
import LoginPage from './pages/auth/LoginPage';
import AdminDashboard from './pages/admin/Dashboard';
import AdminScanPage from './pages/admin/AdminScanPage';
import CourierPage from './pages/Courier/CourierPage';
import * as T from './types';

type TabType = 'Dashboard' | 'Scanning' | 'Orders' | 'Track' | 'History';
type ScreenType = 'Login' | 'Register' | 'Main';

function App() {
  const [currentUser, setCurrentUser] = useState<T.RegisterResponse | null>(null);
  const [currentScreen, setCurrentScreen] = useState<ScreenType>('Login');
  const [activeTab, setActiveTab] = useState<TabType>('Orders');
  const [prepopulatedResi, setPrepopulatedResi] = useState<string>('');

  const handleOrderCreated = (resiId: string) => {
    setPrepopulatedResi(resiId);
    setActiveTab('Track');
  };

  const handleTabChange = (tab: TabType) => {
    if (tab !== 'Track') {
      setPrepopulatedResi('');
    }
    setActiveTab(tab);
  };

  const handleRegisterSuccess = (user: T.RegisterResponse) => {
    localStorage.setItem('token', `mock-register-token-${user.id}`);
    setCurrentUser(user);
    setCurrentScreen('Main');
    setActiveTab('Orders');
  };

  const handleLoginSuccess = (token: string, email: string) => {
    localStorage.setItem('token', token);
    const username = email.split('@')[0];
    const formattedName = username
      .split('.')
      .map((s) => s.charAt(0).toUpperCase() + s.slice(1))
      .join(' ');

    const role = email.toLowerCase().includes('admin')
      ? 'admin'
      : email.toLowerCase().includes('kurir') || email.toLowerCase().includes('courier')
      ? 'courier'
      : 'customer';

    const loggedInUser: T.RegisterResponse = {
      id: Math.floor(1000 + Math.random() * 9000),
      email: email,
      full_name: formattedName || 'User',
      role: role,
      created_at: new Date().toISOString(),
    };

    setCurrentUser(loggedInUser);
    setCurrentScreen('Main');
    setActiveTab(role === 'admin' ? 'Dashboard' : 'Orders');
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setCurrentUser(null);
    setCurrentScreen('Login');
    setPrepopulatedResi('');
  };

  const getInitials = (name: string) => {
    if (!name) return 'U';
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .substring(0, 2);
  };

  // Courier gets their own full-screen portal (no standard shell)
  if (currentScreen === 'Main' && currentUser?.role === 'courier') {
    return <CourierPage />;
  }

  // Screen routing
  if (currentScreen === 'Login') {
    return (
      <LoginPage 
        onLoginSuccess={handleLoginSuccess}
        onNavigateToRegister={() => setCurrentScreen('Register')}
      />
    );
  }

  if (currentScreen === 'Register') {
    return (
      <RegisterPage 
        onRegisterSuccess={handleRegisterSuccess}
        onNavigateToLogin={() => setCurrentScreen('Login')}
      />
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-slate-50/50 text-slate-800">
      
      {/* GLOBAL HEADER (Matching Mockup) */}
      <header className="sticky top-0 bg-white border-b border-slate-100 z-40 px-6 py-4 shadow-sm shadow-slate-100/20">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          
          {/* Logo Section */}
          <div className="flex items-center gap-2.5 cursor-pointer" onClick={() => handleTabChange('Dashboard')}>
            <div className="p-2 bg-gradient-to-tr from-[#2DB7F2] to-[#009ADA] rounded-xl text-white shadow-md shadow-sky-300/30">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-5 h-5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
              </svg>
            </div>
            <span className="text-xl font-black text-slate-900 tracking-tight">
              Sky<span className="text-[#2DB7F2]">Logistics</span>
            </span>
          </div>

          {/* Navigation Tabs */}
          <nav className="hidden md:flex items-center gap-8">
            {(
              currentUser?.role === 'admin'
                ? (['Dashboard', 'Scanning', 'Orders', 'Track', 'History'] as TabType[])
                : (['Dashboard', 'Orders', 'Track', 'History'] as TabType[])
            ).map((tab) => {
              const isActive = activeTab === tab;
              return (
                <button
                  key={tab}
                  onClick={() => handleTabChange(tab)}
                  className={`relative py-1 text-sm font-semibold tracking-wide transition-all duration-200 ${
                    isActive ? 'text-[#2DB7F2]' : 'text-slate-500 hover:text-slate-800'
                  }`}
                >
                  {tab}
                  {isActive && (
                    <span className="absolute bottom-0 left-0 right-0 h-[2.5px] bg-[#2DB7F2] rounded-full animate-fadeIn" />
                  )}
                </button>
              );
            })}
          </nav>

          {/* User & Actions Section */}
          <div className="flex items-center gap-5">
            {/* Notifications */}
            <button className="relative p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded-xl transition-all duration-150">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
              </svg>
              <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-rose-500 rounded-full ring-2 ring-white animate-pulse" />
            </button>

            {/* Help / Support */}
            <button className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded-xl transition-all duration-150">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
              </svg>
            </button>

            {/* User Profile Dropdown */}
            {currentUser && (
              <div className="flex items-center gap-3 border-l border-slate-100 pl-4">
                <div className="w-9 h-9 bg-sky-100 rounded-full overflow-hidden border border-sky-200 flex items-center justify-center font-bold text-[#009ADA]" title={currentUser.email}>
                  {getInitials(currentUser.full_name)}
                </div>
                <div className="hidden sm:flex flex-col text-left">
                  <span className="text-sm font-bold text-slate-700 leading-tight">{currentUser.full_name}</span>
                  <span className="text-[10px] text-slate-400 font-semibold uppercase">{currentUser.role}</span>
                </div>
                <button 
                  onClick={handleLogout}
                  title="Keluar / Logout"
                  className="p-1.5 text-slate-400 hover:text-rose-500 hover:bg-rose-50 rounded-lg transition-all"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4.5 h-4.5">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75" />
                  </svg>
                </button>
              </div>
            )}
          </div>

        </div>
      </header>

      {/* MOBILE TAB BAR */}
      <div className="md:hidden sticky top-[73px] bg-white border-b border-slate-100 z-35 flex items-center justify-around py-2 shadow-sm">
        {(
          currentUser?.role === 'admin'
            ? (['Dashboard', 'Scanning', 'Orders', 'Track', 'History'] as TabType[])
            : (['Dashboard', 'Orders', 'Track', 'History'] as TabType[])
        ).map((tab) => {
          const isActive = activeTab === tab;
          return (
            <button
              key={tab}
              onClick={() => handleTabChange(tab)}
              className={`text-xs font-bold py-1 px-3.5 rounded-full transition-all duration-150 ${
                isActive ? 'bg-sky-50 text-[#2DB7F2]' : 'text-slate-500'
              }`}
            >
              {tab}
            </button>
          );
        })}
      </div>

      {/* MAIN CONTAINER */}
      <main className="flex-grow">
        {currentUser && activeTab === 'Dashboard' && (
          currentUser.role === 'admin' ? (
            <AdminDashboard />
          ) : (
            <div className="max-w-7xl mx-auto py-12 px-6 space-y-8 animate-fadeIn">
              <div>
                <h1 className="text-3xl font-extrabold text-slate-900">Dashboard Utama</h1>
                <p className="text-slate-500 mt-1">Selamat datang kembali, {currentUser.full_name}. Berikut ringkasan aktivitas pengiriman Anda.</p>
              </div>
              
              {/* Status Statistics */}
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
                {[
                  { title: 'Total Pengiriman', count: '12 Paket', color: 'border-slate-200 text-slate-900 bg-white' },
                  { title: 'Dalam Perjalanan', count: '3 Paket', color: 'border-blue-200 text-blue-600 bg-blue-50/20' },
                  { title: 'Berhasil Terkirim', count: '8 Paket', color: 'border-emerald-200 text-emerald-600 bg-emerald-50/20' },
                  { title: 'Gagal Pengiriman', count: '1 Paket', color: 'border-rose-200 text-rose-600 bg-rose-50/20' },
                ].map((stat, i) => (
                  <div key={i} className={`p-6 border rounded-2xl shadow-sm ${stat.color}`}>
                    <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">{stat.title}</h3>
                    <p className="text-3xl font-black mt-2">{stat.count}</p>
                  </div>
                ))}
              </div>

              {/* Quick action card */}
              <div className="p-8 bg-gradient-to-r from-[#2DB7F2] to-[#009ADA] rounded-2xl text-white shadow-xl shadow-sky-300/20 flex flex-col md:flex-row items-center justify-between gap-6">
                <div className="space-y-2 text-center md:text-left">
                  <h3 className="text-2xl font-black">Butuh Pengiriman Baru?</h3>
                  <p className="text-white/80 text-sm max-w-xl">Buat order dan kirim paket Anda ke mana saja dengan cepat menggunakan layanan express maupun regular.</p>
                </div>
                <button
                  onClick={() => handleTabChange('Orders')}
                  className="px-6 py-3.5 bg-white text-[#009ADA] font-extrabold rounded-xl shadow-md active:scale-95 transition-all duration-150"
                >
                  Buat Order Sekarang
                </button>
              </div>
            </div>
          )
        )}

        {activeTab === 'Scanning' && currentUser?.role === 'admin' && (
          <AdminScanPage />
        )}

        {activeTab === 'Orders' && (
          <CreateOrderPage onOrderCreated={handleOrderCreated} />
        )}

        {activeTab === 'Track' && (
          <TrackingPage initialResiID={prepopulatedResi} />
        )}

        {activeTab === 'History' && (
          <div className="max-w-7xl mx-auto py-12 px-6 space-y-8 animate-fadeIn">
            <div>
              <h1 className="text-3xl font-extrabold text-slate-900">Riwayat Pengiriman</h1>
              <p className="text-slate-500 mt-1">Daftar lengkap paket yang telah Anda daftarkan dan dikirim.</p>
            </div>

            <div className="bg-white border border-slate-100 rounded-2xl shadow-md overflow-hidden">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="bg-slate-50 text-slate-400 text-xs font-bold tracking-wider uppercase border-b border-slate-100">
                    <th className="p-4 pl-6">Tanggal</th>
                    <th className="p-4">No. Resi</th>
                    <th className="p-4">Penerima</th>
                    <th className="p-4">Rute</th>
                    <th className="p-4">Status</th>
                    <th className="p-4 text-right pr-6">Aksi</th>
                  </tr>
                </thead>
                <tbody className="text-sm font-semibold text-slate-700 divide-y divide-slate-100">
                  {[
                    { date: '18 Juni 2026', resi: 'RESI-67890', recipient: 'Ani', route: 'Jakarta → Bandung', status: 'DELIVERED', statusColor: 'bg-emerald-50 text-emerald-600 border-emerald-100' },
                    { date: '17 Juni 2026', resi: 'RESI-98124', recipient: 'Budiono', route: 'Jakarta → Surabaya', status: 'IN_TRANSIT', statusColor: 'bg-blue-50 text-blue-600 border-blue-100' },
                    { date: '15 Juni 2026', resi: 'RESI-10395', recipient: 'Siti Sarah', route: 'Jakarta → Jakarta', status: 'CREATED', statusColor: 'bg-sky-50 text-sky-600 border-sky-100' },
                    { date: '10 Juni 2026', resi: 'RESI-29471', recipient: 'Dedi Kurnia', route: 'Jakarta → Surabaya', status: 'FAILED', statusColor: 'bg-rose-50 text-rose-600 border-rose-100' },
                  ].map((row, idx) => (
                    <tr key={idx} className="hover:bg-slate-50/50 transition-colors">
                      <td className="p-4 pl-6 text-slate-500 font-normal">{row.date}</td>
                      <td className="p-4 font-mono font-bold text-slate-900">{row.resi}</td>
                      <td className="p-4">{row.recipient}</td>
                      <td className="p-4 font-normal text-slate-500">{row.route}</td>
                      <td className="p-4">
                        <span className={`px-2.5 py-1 text-xs font-bold rounded-full border ${row.statusColor}`}>{row.status}</span>
                      </td>
                      <td className="p-4 text-right pr-6">
                        <button
                          onClick={() => {
                            setPrepopulatedResi(row.resi);
                            setActiveTab('Track');
                          }}
                          className="text-[#2DB7F2] hover:text-[#009ADA] text-xs font-bold hover:underline"
                        >
                          Detail Lacak
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>

      {/* GLOBAL FOOTER (Matching Mockup) */}
      <footer className="bg-white border-t border-slate-100 py-10 px-6 mt-12">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
          
          {/* Logo & Copyright */}
          <div className="flex flex-col items-center md:items-start text-center md:text-left gap-1">
            <span className="text-lg font-bold text-slate-800 tracking-tight">
              Sky<span className="text-[#2DB7F2]">Logistics</span>
            </span>
            <span className="text-xs text-slate-400 font-normal">
              © 2024 SkyLogistics International. Reliable. Efficient. Optimistic.
            </span>
          </div>

          {/* Policy & Links */}
          <div className="flex items-center gap-6 text-xs font-semibold text-slate-400">
            <a href="#terms" className="hover:text-slate-600 transition-colors">Terms of Service</a>
            <a href="#privacy" className="hover:text-slate-600 transition-colors">Privacy Policy</a>
            <a href="#contact" className="hover:text-slate-600 transition-colors">Contact Support</a>
          </div>

        </div>
      </footer>

    </div>
  );
}

export default App;
