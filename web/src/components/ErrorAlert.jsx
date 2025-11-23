import React from 'react';
import { Alert } from '@mui/material';

/**
 * 可复用的错误提示组件
 * @param {Object} props
 * @param {string} props.message - 错误信息
 * @param {function} [props.onClose] - 关闭错误提示的回调函数
 */
const ErrorAlert = ({ message, onClose }) => {
  if (!message) return null;
  
  return (
    <Alert 
      severity="error" 
      sx={{ mb: 3 }}
      onClose={onClose}
    >
      {message}
    </Alert>
  );
};

export default ErrorAlert;