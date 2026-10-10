import React from 'react'

interface BrandLogoProps {
  size?: number | string
  className?: string
  alt?: string
}

export const BrandLogo: React.FC<BrandLogoProps> = ({ size = 28, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 64 64"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={`shrink-0 select-none ${className}`}
    aria-hidden="true"
  >
    <defs>
      <linearGradient id="brand-bg" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#4F46E5" />
        <stop offset="60%" stopColor="#6366F1" />
        <stop offset="100%" stopColor="#7C3AED" />
      </linearGradient>
      <linearGradient id="brand-highlight" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#FFFFFF" stopOpacity="0.35" />
        <stop offset="100%" stopColor="#FFFFFF" stopOpacity="0.02" />
      </linearGradient>
      <linearGradient id="brand-bolt" x1="20%" y1="0%" x2="80%" y2="100%">
        <stop offset="0%" stopColor="#FFFFFF" />
        <stop offset="60%" stopColor="#F8FAFC" />
        <stop offset="100%" stopColor="#C7D2FE" />
      </linearGradient>
      <linearGradient id="brand-check" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#38BDF8" />
        <stop offset="100%" stopColor="#818CF8" />
      </linearGradient>
      <filter id="brand-shadow" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow dx="0" dy="2" stdDeviation="2.5" floodColor="#1E1B4B" floodOpacity="0.45" />
      </filter>
    </defs>

    {/* Rounded Squircle */}
    <rect x="2" y="2" width="60" height="60" rx="16" fill="url(#brand-bg)" />
    {/* Subtle Inner Highlight Border */}
    <rect
      x="3"
      y="3"
      width="58"
      height="58"
      rx="15"
      fill="none"
      stroke="url(#brand-highlight)"
      strokeWidth="1.5"
    />

    {/* Stylized Lightning Checkmark Emblem */}
    <g filter="url(#brand-shadow)">
      <path d="M35 11L18 33.5H31L26.5 53L46 27H33L38 11H35Z" fill="url(#brand-bolt)" />
      <circle cx="47" cy="46" r="3.5" fill="url(#brand-check)" />
    </g>
  </svg>
)

export default BrandLogo
