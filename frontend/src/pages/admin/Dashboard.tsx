import React, { useState, useEffect } from 'react';
import axios from 'axios';
import * as T from '../../types';

// =============================================================================
// Mock Data Matching the required JSON Response Contract
// =============================================================================
const MOCK_ADMIN_DASHBOARD_DATA = {
  summary: {
    total_orders: 24592,
    packages_delivered: 18205,
    notification_success_rate: 99.4
  },
  hub_weight_analytics: [
    { hub_code: "CGK", total_weight: 1250.8 },
    { hub_code: "BDO", total_weight: 980.4 },
    { hub_code: "SUB", total_weight: 1420.1 },
    { hub_code: "KNO", total_weight: 510.6 },
    { hub_code: "UPG", total_weight: 780.2 }
  ],
  notification_channels: {
    email_percentage: 98.2,
    sms_percentage: 85.4,
    whatsapp_percentage: 92.7
  },
  high_priority_logs: [
    {
      resi_id: "SKY-8829-JKT",
      route: "CGK → SUB",
      timestamp: "2026-06-19T14:32:00+07:00",
      status: "FAILED" as T.TrackingStatus
    },
    {
      resi_id: "SKY-4410-BDO",
      route: "BDO → KNO",
      timestamp: "2026-06-19T13:15:00+07:00",
      status: "RETURNED" as T.TrackingStatus
    },
    {
      resi_id: "SKY-9012-UPG",
      route: "UPG → DPS",
      timestamp: "2026-06-19T11:45:00+07:00",
      status: "IN_TRANSIT" as T.TrackingStatus
    },
    {
      resi_id: "SKY-3391-CGK",
      route: "JOG → CGK",
      timestamp: "2026-06-18T18:20:00+07:00",
      status: "CREATED" as T.TrackingStatus
    }
  ]
};

export const Dashboard: React.FC = () => {
  const [data, setData] = useState(MOCK_ADMIN_DASHBOARD_DATA);
  const [logFilter, setLogFilter] = useState<'ALL' | 'FAILED' | 'RETURNED' | 'IN_TRANSIT' | 'CREATED'>('ALL');
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState('');

  useEffect(() => {
    let active = true;
    const fetchAnalytics = async () => {
      setIsLoading(true);
      setFetchError('');
      
      const token = localStorage.getItem('token');
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      try {
        const response = await axios.get('http://localhost:8081/admin/dashboard-analytics', {
          headers,
        });
        
        if (active) {
          setData(response.data);
          setIsLoading(false);
        }
      } catch (err: any) {
        console.warn('Backend offline or failed to fetch analytics. Falling back to local mock data...', err);
        if (active) {
          setFetchError(err.response?.data?.message || 'Server backend tidak merespons. Berjalan dalam mode simulasi offline.');
          setIsLoading(false);
        }
      }
    };

    fetchAnalytics();
    return () => {
      active = false;
    };
  }, []);

  // Find max weight for bar height ratios
  const maxWeight = Math.max(...(data.hub_weight_analytics || []).map(h => h.total_weight), 1);

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat('en-US').format(num);
  };

  const formatRelativeTime = (isoString: string) => {
    const date = new Date(isoString);
    const now = new Date();
    
    const isToday = date.getDate() === now.getDate() && 
                    date.getMonth() === now.getMonth() && 
                    date.getFullYear() === now.getFullYear();
                    
    const yesterday = new Date(now);
    yesterday.setDate(now.getDate() - 1);
    const isYesterday = date.getDate() === yesterday.getDate() && 
                        date.getMonth() === yesterday.getMonth() && 
                        date.getFullYear() === yesterday.getFullYear();

    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    const timeStr = `${hours}:${minutes}`;

    if (isToday) {
      return `Today, ${timeStr}`;
    } else if (isYesterday) {
      return `Yesterday, ${timeStr}`;
    } else {
      const day = date.getDate();
      const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
      const month = monthNames[date.getMonth()];
      return `${day} ${month}, ${timeStr}`;
    }
  };

  const getStatusBadgeConfig = (status: T.TrackingStatus) => {
    switch (status) {
      case 'FAILED':
        return {
          text: 'CRITICAL',
          bg: 'bg-red-50 border-red-100 text-red-600',
          dot: 'bg-red-500'
        };
      case 'RETURNED':
        return {
          text: 'URGENT',
          bg: 'bg-blue-50 border-blue-100 text-blue-600',
          dot: 'bg-blue-500'
        };
      case 'IN_TRANSIT':
      case 'CREATED':
      default:
        return {
          text: 'INFO',
          bg: 'bg-sky-50/70 border-sky-100/50 text-[#009ADA]',
          dot: 'bg-[#009ADA]'
        };
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-50/50 py-10 px-4 sm:px-6 lg:px-8 space-y-8 animate-pulse">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <div className="space-y-2 w-1/3">
            <div className="h-8 bg-slate-200 rounded-lg w-3/4 animate-pulse"></div>
            <div className="h-4 bg-slate-200 rounded-lg w-1/2 animate-pulse"></div>
          </div>
          <div className="h-8 bg-slate-200 rounded-full w-40 animate-pulse"></div>
        </div>
        <div className="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="h-32 bg-white border border-slate-100 rounded-2xl shadow-sm"></div>
          <div className="h-32 bg-white border border-slate-100 rounded-2xl shadow-sm"></div>
          <div className="h-32 bg-white border border-slate-100 rounded-2xl shadow-sm"></div>
        </div>
        <div className="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 h-80 bg-white border border-slate-100 rounded-2xl shadow-sm"></div>
          <div className="h-80 bg-white border border-slate-100 rounded-2xl shadow-sm"></div>
        </div>
        <div className="max-w-7xl mx-auto h-72 bg-white border border-slate-100 rounded-2xl shadow-sm"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50/50 py-10 px-4 sm:px-6 lg:px-8 space-y-8 animate-fadeIn">
      
      {/* 1. Header Title */}
      <div className="max-w-7xl mx-auto flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-black text-slate-900 tracking-tight">Dashboard Admin</h1>
          <p className="text-slate-500 mt-1">Status analitik logistik, beban Hub, dan performa jaringan SkyLogistics.</p>
        </div>
        
        {/* Status Indicator */}
        <div className="inline-flex items-center gap-2 px-4 py-2 border border-emerald-100 bg-emerald-50 text-emerald-600 rounded-full font-extrabold text-xs">
          <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
          SISTEM OPERASIONAL 100%
        </div>
      </div>

      <div className="max-w-7xl mx-auto space-y-8">
        
        {/* Error/Offline Alert Banner */}
        {fetchError && (
          <div className="p-4 bg-amber-50 border border-amber-100 rounded-xl flex items-start gap-3 text-amber-700 animate-fadeIn">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5 mt-0.5 flex-shrink-0">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div className="text-xs sm:text-sm">
              <span className="font-bold">Mode Simulasi Offline: </span>
              {fetchError}
            </div>
          </div>
        )}

        {/* 2. STAT CARDS GRID (Icon on Left) */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          
          {/* Card 1: Total Orderan */}
          <div className="bg-white border border-sky-100/50 p-6 rounded-2xl shadow-md shadow-sky-100/40 hover:shadow-lg transition-all duration-300 flex items-center gap-5">
            <div className="p-3 bg-sky-50 text-[#2DB7F2] rounded-2xl border border-sky-100/50">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.2} stroke="currentColor" className="w-7 h-7">
                <path strokeLinecap="round" strokeLinejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
              </svg>
            </div>
            <div className="space-y-0.5">
              <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">TOTAL ORDERAN</h3>
              <p className="text-3xl font-black text-slate-800 leading-none">{formatNumber(data.summary.total_orders)}</p>
            </div>
          </div>

          {/* Card 2: Paket Terkirim */}
          <div className="bg-white border border-sky-100/50 p-6 rounded-2xl shadow-md shadow-sky-100/40 hover:shadow-lg transition-all duration-300 flex items-center gap-5">
            <div className="p-3 bg-emerald-50 text-emerald-500 rounded-2xl border border-emerald-100/50">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.2} stroke="currentColor" className="w-7 h-7">
                <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.113c0-.732-.412-1.4-1.071-1.724a8.256 8.256 0 00-3.03-.69c-.429-.027-.854-.01-1.272.049-.665.093-1.178.675-1.178 1.35v.193m2.26 12.015h-.008v-.008h.008v.008zm-1.22 0h-.008v-.008h.008v.008z" />
              </svg>
            </div>
            <div className="space-y-0.5">
              <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">PAKET TERKIRIM</h3>
              <p className="text-3xl font-black text-slate-800 leading-none">{formatNumber(data.summary.packages_delivered)}</p>
            </div>
          </div>

          {/* Card 3: Notifikasi Sukses */}
          <div className="bg-white border border-sky-100/50 p-6 rounded-2xl shadow-md shadow-sky-100/40 hover:shadow-lg transition-all duration-300 flex items-center gap-5">
            <div className="p-3 bg-sky-50 text-[#2DB7F2] rounded-2xl border border-sky-100/50">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.2} stroke="currentColor" className="w-7 h-7">
                <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
              </svg>
            </div>
            <div className="space-y-0.5">
              <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">NOTIFIKASI SUKSES</h3>
              <p className="text-3xl font-black text-slate-800 leading-none">{data.summary.notification_success_rate}%</p>
            </div>
          </div>

        </div>

        {/* 3. CHART & PROGRESS BARS SECTION */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          
          {/* Bar Chart Column (With Capsule Wrapper) */}
          <div className="lg:col-span-2 bg-white border border-sky-100/50 p-6 rounded-2xl shadow-md shadow-sky-100/40 flex flex-col justify-between">
            <div className="flex items-center justify-between mb-8">
              <h2 className="text-lg font-bold text-slate-800">Beban Berat Paket Per HUB Asal</h2>
              
              {/* Three dots menu */}
              <button className="text-slate-400 hover:text-slate-600 transition-colors p-1 rounded hover:bg-slate-50">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-5 h-5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM12.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM18.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0z" />
                </svg>
              </button>
            </div>

            {/* Custom Capsule Bar Chart */}
            <div className="relative flex items-end justify-around h-64 border-b border-slate-100/80 pb-2">
              {data.hub_weight_analytics.map((hub, i) => {
                const ratio = (hub.total_weight / maxWeight) * 100; // Filled bar percentage inside capsule
                return (
                  <div key={i} className="flex flex-col items-center group relative w-12 sm:w-16">
                    
                    {/* Weight value tooltip on hover */}
                    <span className="absolute -top-10 scale-0 group-hover:scale-100 transition-all duration-150 bg-slate-800 text-white text-[10px] font-bold px-2 py-1 rounded shadow-md whitespace-nowrap z-10">
                      {formatNumber(hub.total_weight)} Kg
                    </span>

                    {/* Capsule bar container */}
                    <div className="w-8 sm:w-10 h-48 bg-sky-50/70 border border-sky-100/20 rounded-t-lg relative overflow-hidden flex flex-col justify-end">
                      {/* Inner Filled Bar */}
                      <div 
                        style={{ height: `${ratio}%` }}
                        className="w-full bg-gradient-to-t from-[#009ADA] to-[#2DB7F2] rounded-t-lg transition-all duration-500 hover:opacity-90 shadow-md shadow-sky-100/30"
                      />
                    </div>

                    {/* Weight label static */}
                    <span className="text-[10px] font-extrabold text-[#009ADA] mt-1.5">
                      {Math.round(hub.total_weight)} Kg
                    </span>
                  </div>
                );
              })}
            </div>

            {/* X-axis Hub names */}
            <div className="flex justify-around pt-3 text-xs font-bold text-slate-500">
              {data.hub_weight_analytics.map((hub, i) => (
                <span key={i} className="w-12 sm:w-16 text-center">{hub.hub_code}</span>
              ))}
            </div>
          </div>

          {/* Notification Status Column */}
          <div className="bg-white border border-sky-100/50 p-6 rounded-2xl shadow-md shadow-sky-100/40 flex flex-col justify-between space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-bold text-slate-800">Status Pengiriman Notifikasi</h2>
                <p className="text-xs text-slate-400 mt-0.5 font-medium">Rasio sukses pengiriman pesan per saluran.</p>
              </div>
              <div className="text-slate-400">
                {/* Megaphone Icon */}
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.114 5.636a9 9 0 010 12.728M16.463 8.288a5.25 5.25 0 010 7.424M6.75 8.25l4.72-4.72a.75.75 0 011.28.53v15.88a.75.75 0 01-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.01 9.01 0 012.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75z" />
                </svg>
              </div>
            </div>

            {/* Progress Bars */}
            <div className="space-y-6 flex-grow justify-center flex flex-col pb-4">
              {/* Email */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between text-xs font-bold">
                  <span className="text-slate-600 flex items-center gap-1.5">
                    {/* Email Icon */}
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4 text-slate-400">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
                    </svg>
                    Email
                  </span>
                  <span className="text-slate-700">{data.notification_channels.email_percentage}%</span>
                </div>
                <div className="h-2.5 w-full bg-slate-50 border border-slate-100/50 rounded-full overflow-hidden">
                  <div 
                    style={{ width: `${data.notification_channels.email_percentage}%` }}
                    className="h-full bg-[#006F9E] rounded-full transition-all duration-500"
                  />
                </div>
              </div>

              {/* SMS */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between text-xs font-bold">
                  <span className="text-slate-600 flex items-center gap-1.5">
                    {/* Device Icon */}
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4 text-slate-400">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3" />
                    </svg>
                    SMS
                  </span>
                  <span className="text-slate-700">{data.notification_channels.sms_percentage}%</span>
                </div>
                <div className="h-2.5 w-full bg-slate-50 border border-slate-100/50 rounded-full overflow-hidden">
                  <div 
                    style={{ width: `${data.notification_channels.sms_percentage}%` }}
                    className="h-full bg-[#2DB7F2] rounded-full transition-all duration-500"
                  />
                </div>
              </div>

              {/* WhatsApp */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between text-xs font-bold">
                  <span className="text-slate-600 flex items-center gap-1.5">
                    {/* Chat Icon */}
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4 text-slate-400">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.322-.047-.647-.285-.873C2.85 16.82 1.5 14.54 1.5 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" />
                    </svg>
                    WhatsApp API
                  </span>
                  <span className="text-slate-700">{data.notification_channels.whatsapp_percentage}%</span>
                </div>
                <div className="h-2.5 w-full bg-slate-50 border border-slate-100/50 rounded-full overflow-hidden">
                  <div 
                    style={{ width: `${data.notification_channels.whatsapp_percentage}%` }}
                    className="h-full bg-emerald-500 rounded-full transition-all duration-500"
                  />
                </div>
              </div>
            </div>

          </div>

        </div>

        {/* 4. HIGH-PRIORITY LOGS TABLE */}
        <div className="bg-white border border-sky-100/50 p-6 rounded-2xl shadow-md shadow-sky-100/40 space-y-5">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-bold text-slate-800">High-Priority Logistics Log</h2>
              <p className="text-xs text-slate-400 mt-0.5">Filter status pengiriman prioritas di bawah ini.</p>
            </div>
            
            {/* Filter pills and view all action */}
            <div className="flex flex-wrap items-center gap-3 text-xs font-bold">
              <div className="flex gap-2">
                {(['ALL', 'FAILED', 'IN_TRANSIT', 'RETURNED', 'CREATED'] as const).map((status) => {
                  const isActive = logFilter === status;
                  return (
                    <button
                      key={status}
                      type="button"
                      onClick={() => setLogFilter(status)}
                      className={`px-3 py-1.5 rounded-lg border transition-all duration-150 ${
                        isActive
                          ? 'bg-[#2DB7F2] border-[#2DB7F2] text-white shadow-sm shadow-sky-100'
                          : 'bg-slate-50 border-slate-100 hover:bg-slate-100 text-slate-500'
                      }`}
                    >
                      {status === 'ALL' ? 'ALL' : getStatusBadgeConfig(status).text}
                    </button>
                  );
                })}
              </div>
              
              <button className="text-xs font-bold text-[#009ADA] hover:text-[#2DB7F2] transition-colors flex items-center gap-1 ml-2">
                <span>View All</span>
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-3.5 h-3.5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
                </svg>
              </button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-slate-50/70 text-slate-400 text-xs font-bold tracking-wider uppercase border-b border-slate-100">
                  <th className="p-4 pl-6">AWB ID</th>
                  <th className="p-4">Route</th>
                  <th className="p-4">Timestamp</th>
                  <th className="p-4 pr-6">Status</th>
                </tr>
              </thead>
              <tbody className="text-sm font-semibold text-slate-700 divide-y divide-slate-100">
                {data.high_priority_logs
                  .filter((log) => logFilter === 'ALL' || log.status === logFilter)
                  .map((log, idx) => {
                    const badge = getStatusBadgeConfig(log.status);
                    return (
                      <tr key={idx} className="hover:bg-slate-50/50 transition-colors">
                        {/* Resi ID */}
                        <td className="p-4 pl-6 font-mono font-bold text-slate-900 flex items-center gap-1.5">
                          {log.resi_id}
                          <button
                            onClick={() => navigator.clipboard.writeText(log.resi_id)}
                            className="p-1 hover:bg-slate-50 border border-slate-100 rounded text-slate-400 hover:text-slate-600 transition-colors"
                            title="Salin Resi"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-3 h-3">
                              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 7.5V6.108c0-1.135.845-2.098 1.976-2.192.373-.03.748-.057 1.123-.08M15.75 18H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08M15.75 18.75v-1.875a3.375 3.375 0 00-3.375-3.375h-1.5a1.125 1.125 0 01-1.125-1.125v-1.5A3.375 3.375 0 006.375 7.5H5.25m11.9-3.664A2.251 2.251 0 0015 2.25h-1.5a2.251 2.251 0 00-2.15 1.586m5.8 0c.065.21.1.433.1.664v.75h-6V4.5c0-.231.035-.454.1-.664M6.75 7.5H4.875c-.621 0-1.125.504-1.125 1.125v12c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V16.5a9 9 0 00-9-9z" />
                            </svg>
                          </button>
                        </td>
                        
                        {/* Route */}
                        <td className="p-4 font-normal text-slate-500">{log.route}</td>
                        
                        {/* Timestamp */}
                        <td className="p-4 font-normal text-slate-400">
                          {formatRelativeTime(log.timestamp)}
                        </td>
                        
                        {/* Status */}
                        <td className="p-4 pr-6">
                          <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full border text-xs font-bold ${badge.bg}`}>
                            <span className={`w-1.5 h-1.5 rounded-full ${badge.dot}`} />
                            {badge.text}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>
        </div>

      </div>
    </div>
  );
};
export default Dashboard;
