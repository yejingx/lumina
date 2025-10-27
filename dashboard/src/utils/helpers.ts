import { message } from 'antd';
import dayjs from 'dayjs';
import { DATE_FORMAT } from './constants';

// 格式化日期
export const formatDate = (date: string | undefined, format: string = DATE_FORMAT): string => {
  if (!date) return '-';
  return dayjs(date).format(format);
};

// 判断某时间是否超过指定分钟数
export const isOlderThanMinutes = (date: string | undefined, minutes: number): boolean => {
  if (!date) return true;
  const time = dayjs(date);
  if (!time.isValid()) return true;
  const now = dayjs();
  const diff = now.diff(time, 'minute');
  return diff >= minutes;
};

// 处理 API 错误
export const handleApiError = (error: any, defaultMessage: string = '操作失败') => {
  console.error('API Error:', error);
  const errorMessage = error?.response?.data?.message || error?.message || defaultMessage;
  message.error(errorMessage);
};

// 确认删除对话框配置
export const getDeleteConfirmConfig = (title: string, content?: string) => ({
  title,
  content: content || '删除后无法恢复，确定要删除吗？',
  okText: '确定',
  okType: 'danger' as const,
  cancelText: '取消',
});

// 生成随机 ID
export const generateId = (): string => {
  return Math.random().toString(36).substr(2, 9);
};

// 防抖函数
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

// 节流函数
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean;
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
};

// 复制到剪贴板
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    await navigator.clipboard.writeText(text);
    message.success('已复制到剪贴板');
    return true;
  } catch (err) {
    console.error('复制失败:', err);
    message.error('复制失败');
    return false;
  };
};

// 下载文件
export const downloadFile = (url: string, filename: string) => {
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

// 获取文件大小格式化字符串
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};