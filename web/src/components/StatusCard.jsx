import React from 'react';
import { Card, CardContent, Typography } from '@mui/material';

/**
 * 可复用的状态卡片组件
 * @param {Object} props
 * @param {string} props.title - 卡片标题
 * @param {string|number} props.value - 卡片显示的值
 * @param {string} [props.color] - 文本颜色，可以是MUI颜色值
 * @param {string} [props.bgColor] - 卡片背景色
 */
const StatusCard = ({ title, value, color, bgColor = '#2c3e50' }) => {
  return (
    <Card sx={{ backgroundColor: bgColor }}>
      <CardContent>
        <Typography variant="h6" color="text.secondary" gutterBottom>
          {title}
        </Typography>
        <Typography variant="h4" color={color}>
          {value}
        </Typography>
      </CardContent>
    </Card>
  );
};

export default StatusCard;