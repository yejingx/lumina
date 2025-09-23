import React, { useState, useEffect, useRef } from 'react';
import { PlayCircleOutlined, FileImageOutlined, CloseOutlined } from '@ant-design/icons';
import { Modal } from 'antd';

interface VideoThumbnailProps {
  videoUrl: string;
  width?: number;
  height?: number;
  onClick?: () => void;
  className?: string;
  style?: React.CSSProperties;
  title?: string;
  enablePreview?: boolean; // New prop to enable/disable modal preview
}

const VideoThumbnail: React.FC<VideoThumbnailProps> = ({
  videoUrl,
  width = 100,
  height = 100,
  onClick,
  className,
  style,
  title = '点击播放视频',
  enablePreview = false
}) => {
  const [thumbnailUrl, setThumbnailUrl] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [previewVisible, setPreviewVisible] = useState(false);
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!videoUrl) {
      setLoading(false);
      setError(true);
      return;
    }

    const extractThumbnail = () => {
      const video = videoRef.current;
      const canvas = canvasRef.current;
      
      if (!video || !canvas) return;

      const context = canvas.getContext('2d');
      if (!context) return;

      // Set canvas dimensions
      canvas.width = width;
      canvas.height = height;

      // Draw the current frame to canvas
      context.drawImage(video, 0, 0, width, height);

      // Convert canvas to data URL
      const dataURL = canvas.toDataURL('image/jpeg', 0.8);
      setThumbnailUrl(dataURL);
      setLoading(false);
    };

    const handleLoadedData = () => {
      const video = videoRef.current;
      if (!video) return;

      // Seek to first frame (0.1 seconds to ensure we get a frame)
      video.currentTime = 0.1;
    };

    const handleSeeked = () => {
      extractThumbnail();
    };

    const handleError = () => {
      setError(true);
      setLoading(false);
    };

    const video = videoRef.current;
    if (video) {
      video.addEventListener('loadeddata', handleLoadedData);
      video.addEventListener('seeked', handleSeeked);
      video.addEventListener('error', handleError);
      
      // Set video source
      video.src = videoUrl;
      video.load();

      return () => {
        video.removeEventListener('loadeddata', handleLoadedData);
        video.removeEventListener('seeked', handleSeeked);
        video.removeEventListener('error', handleError);
      };
    }
  }, [videoUrl, width, height]);

  const containerStyle: React.CSSProperties = {
    width,
    height,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#f5f5f5',
    border: '1px solid #d9d9d9',
    borderRadius: '6px',
    cursor: onClick ? 'pointer' : 'default',
    position: 'relative',
    overflow: 'hidden',
    ...style
  };

  const overlayStyle: React.CSSProperties = {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    fontSize: '24px',
    color: 'rgba(255, 255, 255, 0.8)',
    textShadow: '0 0 4px rgba(0, 0, 0, 0.5)',
    zIndex: 2
  };

  const thumbnailStyle: React.CSSProperties = {
    width: '100%',
    height: '100%',
    objectFit: 'cover'
  };

  const handleClick = () => {
    if (enablePreview) {
      setPreviewVisible(true);
    } else if (onClick) {
      onClick();
    }
  };

  return (
    <>
      <div
        className={className}
        style={containerStyle}
        onClick={handleClick}
        title={title}
      >
        {/* Hidden video element for thumbnail extraction */}
        <video
          ref={videoRef}
          style={{ display: 'none' }}
          muted
          preload="metadata"
          crossOrigin="anonymous"
        />
        
        {/* Hidden canvas for thumbnail generation */}
        <canvas
          ref={canvasRef}
          style={{ display: 'none' }}
        />

        {loading && (
          <>
            <FileImageOutlined style={{ fontSize: '24px', color: '#bfbfbf' }} />
            <div style={{ position: 'absolute', bottom: '4px', right: '4px', fontSize: '12px', color: '#bfbfbf' }}>
              加载中...
            </div>
          </>
        )}

        {error && !loading && (
          <>
            <FileImageOutlined style={{ fontSize: '24px', color: '#bfbfbf' }} />
            <PlayCircleOutlined style={overlayStyle} />
          </>
        )}

        {thumbnailUrl && !loading && !error && (
          <>
            <img
              src={thumbnailUrl}
              alt="Video thumbnail"
              style={thumbnailStyle}
            />
            <PlayCircleOutlined style={overlayStyle} />
          </>
        )}
      </div>

      {/* Video Preview Modal */}
      <Modal
        open={previewVisible}
        onCancel={() => setPreviewVisible(false)}
        footer={null}
        width="80%"
        style={{ top: 20 }}
        bodyStyle={{ padding: 0 }}
        destroyOnClose
      >
        <video
          controls
          autoPlay
          style={{
            width: '100%',
            height: 'auto',
            maxHeight: '80vh',
            display: 'block'
          }}
          src={videoUrl}
        >
          您的浏览器不支持视频播放。
        </video>
      </Modal>
    </>
  );
};

export default VideoThumbnail;