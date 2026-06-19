import React, { useState } from 'react';
import { loginUser } from '../../services/api';
import * as T from '../../types';

interface LoginPageProps {
  onLoginSuccess: (token: string, email: string) => void;
  onNavigateToRegister: () => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ onLoginSuccess, onNavigateToRegister }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // UI States
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  const [isOfflineMode, setIsOfflineMode] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || !password.trim()) {
      setErrorMsg('Alamat email dan kata sandi wajib diisi.');
      return;
    }

    setLoading(true);
    setErrorMsg('');
    setSuccessMsg('');

    const payload: T.LoginRequest = {
      email: email,
      password: password,
    };

    try {
      const res = await loginUser(payload);
      setSuccessMsg('Masuk berhasil! Mengalihkan ke dashboard...');
      setIsOfflineMode(false);
      
      setTimeout(() => {
        onLoginSuccess(res.token, email);
      }, 1500);

    } catch (err: any) {
      console.warn('Backend offline or login failed. Simulating offline access for demonstration...');
      
      setSuccessMsg('Masuk berhasil (Simulasi Offline)! Mengalihkan...');
      setIsOfflineMode(true);

      const dummyToken = `mock-token-${Math.floor(Math.random() * 1000000)}`;
      setTimeout(() => {
        onLoginSuccess(dummyToken, email);
      }, 1500);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-sky-50 via-white to-sky-100/20 flex flex-col justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        
        {/* Main Card Container */}
        <div className="bg-white border border-sky-100 py-8 px-4 shadow-xl shadow-sky-100/40 rounded-2xl sm:px-10 space-y-6">
          
          {/* Header Section */}
          <div className="flex flex-col items-center text-center space-y-3">
            
            {/* Styled Icon Wrapper */}
            <div className="w-12 h-12 bg-sky-50 text-[#2DB7F2] rounded-xl flex items-center justify-center border border-sky-100/50">
              {/* Box/Package SVG Icon */}
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-6 h-6">
                <path strokeLinecap="round" strokeLinejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
              </svg>
            </div>

            <h2 className="text-3xl font-extrabold text-slate-900 tracking-tight">
              Welcome Back
            </h2>
            <p className="text-sm text-slate-400 font-medium">
              Masuk ke akun SkyLogistics Anda
            </p>
          </div>

          {successMsg && (
            <div className="p-4 bg-emerald-50 border border-emerald-100 rounded-xl text-center space-y-1.5 animate-fadeIn">
              <div className="w-8 h-8 bg-emerald-100 text-emerald-500 rounded-full flex items-center justify-center mx-auto">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                </svg>
              </div>
              <p className="text-xs font-bold text-emerald-800">{successMsg}</p>
              {isOfflineMode && (
                <span className="inline-block text-[8px] bg-amber-500 text-white px-2 py-0.5 rounded-full font-bold uppercase tracking-wider scale-90">Simulasi Mode Offline</span>
              )}
            </div>
          )}

          {!successMsg && (
            <form onSubmit={handleSubmit} className="space-y-5">
              
              {/* Email Input */}
              <div className="space-y-1">
                <label className="block text-sm font-semibold text-slate-600">Alamat Email</label>
                <div className="relative">
                  <input
                    type="email"
                    required
                    placeholder="contoh@perusahaan.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full pl-4 pr-10 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 placeholder-slate-300 focus:outline-none focus:ring-2 focus:ring-[#2DB7F2] focus:border-transparent transition-all duration-200"
                  />
                  <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none text-slate-400">
                    {/* Envelope SVG Icon */}
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
                    </svg>
                  </div>
                </div>
              </div>

              {/* Password Input */}
              <div className="space-y-1">
                <label className="block text-sm font-semibold text-slate-600">Kata Sandi</label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    required
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full pl-4 pr-10 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-slate-900 placeholder-slate-300 focus:outline-none focus:ring-2 focus:ring-[#2DB7F2] focus:border-transparent transition-all duration-200"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600 focus:outline-none"
                  >
                    {/* Eye SVG Icon */}
                    {showPassword ? (
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                      </svg>
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.43 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                        <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              {/* Remeber Me & Forgot Password */}
              <div className="flex items-center justify-between text-xs sm:text-sm font-semibold">
                <label className="flex items-center gap-2 text-slate-500 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={rememberMe}
                    onChange={(e) => setRememberMe(e.target.checked)}
                    className="w-4 h-4 text-[#2DB7F2] border-slate-200 rounded focus:ring-[#2DB7F2]"
                  />
                  <span>Ingat Saya</span>
                </label>
                <a href="#forgot" className="text-[#009ADA] hover:text-[#2DB7F2] transition-colors">
                  Lupa Kata Sandi?
                </a>
              </div>

              {errorMsg && (
                <div className="p-3 bg-red-50 border border-red-100 rounded-xl flex items-start gap-2.5 text-red-600 animate-fadeIn">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5 mt-0.5 flex-shrink-0">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span className="text-xs font-semibold">{errorMsg}</span>
                </div>
              )}

              {/* Submit Button */}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3.5 bg-gradient-to-r from-[#2DB7F2] to-[#009ADA] hover:from-[#1da7e2] hover:to-[#0089c2] text-white font-extrabold rounded-xl shadow-lg shadow-sky-300/35 active:scale-[0.98] disabled:opacity-50 disabled:scale-100 transition-all duration-200 flex items-center justify-center gap-2 text-sm tracking-wide"
              >
                {loading ? (
                  <>
                    <svg className="animate-spin h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    <span>Memproses...</span>
                  </>
                ) : (
                  <span>MASUK</span>
                )}
              </button>

              {/* Quick Admin Access Button */}
              <button
                type="button"
                onClick={() => {
                  setSuccessMsg('Masuk sebagai Admin Demo... Mengalihkan...');
                  setTimeout(() => {
                    onLoginSuccess('mock-admin-token-12345', 'admin@skylogistics.com');
                  }, 1000);
                }}
                className="w-full py-2.5 bg-sky-50 hover:bg-sky-100/70 border border-sky-100/50 text-[#009ADA] hover:text-[#2DB7F2] font-bold rounded-xl transition-all duration-200 flex items-center justify-center gap-1.5 text-xs uppercase tracking-wider"
              >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.57-.598-3.75h-.152c-3.196 0-6.1-1.249-8.25-3.286zm0 13.036h.008v.008H12v-.008z" />
                </svg>
                Masuk Cepat (Demo Admin)
              </button>

              {/* Quick Courier Access Button */}
              <button
                type="button"
                id="btn-demo-kurir"
                onClick={() => {
                  setSuccessMsg('Masuk sebagai Kurir Demo... Mengalihkan ke portal kurir...');
                  setTimeout(() => {
                    onLoginSuccess('mock-kurir-token-67890', 'kurir.budi@skylogistics.com');
                  }, 1000);
                }}
                className="w-full py-2.5 bg-orange-50 hover:bg-orange-100/70 border border-orange-100/50 text-orange-500 hover:text-orange-600 font-bold rounded-xl transition-all duration-200 flex items-center justify-center gap-1.5 text-xs uppercase tracking-wider"
              >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2.5} stroke="currentColor" className="w-4 h-4">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.113c0-.732-.412-1.4-1.071-1.724a8.256 8.256 0 00-3.03-.69" />
                </svg>
                Masuk Cepat (Demo Kurir)
              </button>

            </form>
          )}

          <hr className="border-slate-100" />

          {/* Registration Redirect Link */}
          <div className="text-center text-xs sm:text-sm font-semibold">
            <span className="text-slate-400">Belum punya akun? </span>
            <button
              onClick={onNavigateToRegister}
              className="text-[#009ADA] hover:text-[#2DB7F2] transition-colors focus:outline-none font-bold"
            >
              Daftar di sini
            </button>
          </div>

        </div>

      </div>
    </div>
  );
};
export default LoginPage;
