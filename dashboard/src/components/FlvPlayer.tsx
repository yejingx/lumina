import React, { useEffect, useRef } from 'react';

// 使用 require 导入 flv.js 避免 TypeScript 类型声明问题
const flvjs = require('flv.js');

// 使用 flv.js 的 http-flv 播放器，兼容主流浏览器（基于 MSE）

interface FlvPlayerProps {
  url: string;
  autoPlay?: boolean;
  controls?: boolean;
  muted?: boolean;
  style?: React.CSSProperties;
  onError?: (type?: any, details?: any) => void;
  onPlaying?: () => void;
}

const FlvPlayer: React.FC<FlvPlayerProps> = ({
  url,
  autoPlay = true,
  controls = true,
  muted = false,
  style,
  onError,
  onPlaying,
}) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const playerRef = useRef<any>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const handlePlaying = () => {
      try {
        onPlaying && onPlaying();
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('onPlaying 回调执行失败:', e);
      }
    };

    const handleVideoError = () => {
      try {
        onError && onError();
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('onError 回调执行失败:', e);
      }
    };

    if (flvjs.isSupported()) {
      const player = flvjs.createPlayer(
        {
          type: 'flv',
          url,
          isLive: true,
          cors: true,
        },
        {
          // 在使用打包环境（如 CRA）时启用 worker 可能导致运行时错误
          // 如“Class extends value undefined is not a constructor or null”。
          // 关闭 worker 以避免该问题。
          enableWorker: false,
          stashInitialSize: 128,
          autoCleanupSourceBuffer: true,
          lazyLoad: true,
          lazyLoadMaxDuration: 30,
          seekType: 'range',
        }
      );
      // 捕获 flv.js 播放器错误，便于定位问题与触发重试
      try {
        player.on(flvjs.Events.ERROR, (type: any, details: any) => {
          // eslint-disable-next-line no-console
          console.error('flv.js 播放器错误:', type, details);
          try {
            onError && onError(type, details);
          } catch (e) {
            // eslint-disable-next-line no-console
            console.warn('onError 回调执行失败:', e);
          }
        });
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('flv.js 事件绑定失败:', e);
      }
      player.attachMediaElement(video);
      player.load();
      if (autoPlay) {
        const playPromise = video.play();
        // 某些浏览器返回 Promise，某些返回 void；做兼容处理
        if (playPromise && typeof (playPromise as any).catch === 'function') {
          (playPromise as any).catch(() => {
            // eslint-disable-next-line no-console
            console.warn('自动播放失败，用户交互后再播放', url);
          });
        }
      }
      // 监听 HTMLVideoElement 的 playing / error 事件
      try {
        video.addEventListener('playing', handlePlaying);
        video.addEventListener('error', handleVideoError);
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('video 事件绑定失败:', e);
      }
      playerRef.current = player;
    } else {
      // eslint-disable-next-line no-console
      console.error('当前环境不支持 flv.js');
    }

    return () => {
      try {
        if (playerRef.current) {
          playerRef.current.unload();
          playerRef.current.destroy();
          playerRef.current = null;
        }
        if (video) {
          video.removeEventListener('playing', handlePlaying);
          video.removeEventListener('error', handleVideoError);
        }
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('flv.js 清理失败:', e);
      }
    };
  }, [url, autoPlay]);

  return (
    <video
      ref={videoRef}
      style={{ width: '100%', height: 'auto', ...style }}
      controls={controls}
      muted={muted}
    />
  );
};

export default FlvPlayer;