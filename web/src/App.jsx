import React, { useState, useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout.jsx'
import Login from './pages/Login.jsx'
import Register from './pages/Register.jsx'
import Dashboard from './pages/Dashboard.jsx'
import Networks from './pages/Networks.jsx'
import NetworkDetail from './pages/NetworkDetail.jsx'
import Members from './pages/Members.jsx'
import Profile from './pages/Profile.jsx'
import Settings from './pages/Settings.jsx'
import api from './services/api.js'
import './App.css'

function App() {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  // 检查用户登录状态
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const token = localStorage.getItem('token')
        if (token) {
          api.defaults.headers.common['Authorization'] = `Bearer ${token}`
          const response = await api.get('/profile')
          setUser(response.data)
        }
      } catch (error) {
        localStorage.removeItem('token')
        delete api.defaults.headers.common['Authorization']
      } finally {
        setLoading(false)
      }
    }

    checkAuth()
  }, [])

  // 处理登录成功
  const handleLoginSuccess = (userData) => {
    setUser(userData)
  }
  
  // 处理注册成功
  const handleRegisterSuccess = () => {
    // 可以在这里添加注册成功后的逻辑
    console.log('注册成功')
  }

  // 处理登出
  const handleLogout = () => {
    setUser(null)
    localStorage.removeItem('token')
    delete api.defaults.headers.common['Authorization']
  }

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        fontSize: '18px'
      }}>
        加载中...
      </div>
    )
  }

  return (
    <div className="app">
      <Routes>
        <Route path="/login" element={<Login onLoginSuccess={handleLoginSuccess} />}></Route>
        <Route path="/register" element={<Register onRegisterSuccess={handleRegisterSuccess} />}></Route>
        
        {user ? (
          <>
            <Route path="/" element={<Layout user={user} onLogout={handleLogout} />}>
              <Route index element={<Dashboard />}></Route>
              <Route path="networks" element={<Networks />}></Route>
              <Route path="networks/:id" element={<NetworkDetail />}></Route>
              <Route path="networks/:id/members" element={<Members />}></Route>
              <Route path="profile" element={<Profile user={user} />}></Route>
              <Route path="settings" element={<Settings />}></Route>
            </Route>
          </>
        ) : (
          <>
            <Route path="/*" element={<Navigate to="/login" replace />}></Route>
          </>
        )}
      </Routes>
    </div>
  )
}

export default App