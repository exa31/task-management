import React, { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

type TooltipPosition = 'top' | 'bottom' | 'left' | 'right'

interface TooltipData {
  text: string
  position: TooltipPosition
  rect: DOMRect
}

/**
 * Universal Event-Delegated Tooltip Component.
 * Automatically displays tooltips for ANY static or dynamically mounted DOM element
 * possessing `data-pr-tooltip` or `title` attributes across modals, drawers, and tables.
 */
export const GlobalTooltip: React.FC = () => {
  const [data, setData] = useState<TooltipData | null>(null)
  const [coords, setCoords] = useState<{ x: number; y: number }>({ x: 0, y: 0 })
  const [actualPos, setActualPos] = useState<TooltipPosition>('top')
  const tooltipRef = useRef<HTMLDivElement>(null)
  const timerRef = useRef<number | null>(null)
  const currentTargetRef = useRef<HTMLElement | null>(null)

  useEffect(() => {
    const handleMouseOver = (e: MouseEvent) => {
      const target = (e.target as HTMLElement | null)?.closest?.(
        '[data-pr-tooltip], [title]'
      ) as HTMLElement | null
      if (!target) return

      let text = target.getAttribute('data-pr-tooltip')
      if (!text) {
        text = target.getAttribute('title')
        if (text) {
          // Suppress browser default title tooltip to prevent dual popups
          target.setAttribute('data-original-title', text)
          target.removeAttribute('title')
        }
      }

      if (!text || !text.trim()) return

      if (currentTargetRef.current === target && data) return
      currentTargetRef.current = target

      if (timerRef.current) {
        window.clearTimeout(timerRef.current)
      }

      timerRef.current = window.setTimeout(() => {
        if (!target.isConnected) return
        const rect = target.getBoundingClientRect()
        const position = (target.getAttribute('data-pr-position') || 'top') as TooltipPosition

        setData({
          text: text.trim(),
          position,
          rect
        })
      }, 120)
    }

    const handleMouseOut = (e: MouseEvent) => {
      const target = currentTargetRef.current
      if (!target) return

      const related = e.relatedTarget as Node | null
      if (related && target.contains(related)) {
        return
      }

      const originalTitle = target.getAttribute('data-original-title')
      if (originalTitle) {
        target.setAttribute('title', originalTitle)
        target.removeAttribute('data-original-title')
      }

      if (timerRef.current) {
        window.clearTimeout(timerRef.current)
        timerRef.current = null
      }
      currentTargetRef.current = null
      setData(null)
    }

    const handleDismiss = () => {
      if (timerRef.current) {
        window.clearTimeout(timerRef.current)
        timerRef.current = null
      }
      if (currentTargetRef.current) {
        const originalTitle = currentTargetRef.current.getAttribute('data-original-title')
        if (originalTitle) {
          currentTargetRef.current.setAttribute('title', originalTitle)
          currentTargetRef.current.removeAttribute('data-original-title')
        }
      }
      currentTargetRef.current = null
      setData(null)
    }

    document.addEventListener('mouseover', handleMouseOver, { passive: true })
    document.addEventListener('mouseout', handleMouseOut, { passive: true })
    window.addEventListener('scroll', handleDismiss, { passive: true, capture: true })
    window.addEventListener('click', handleDismiss, { passive: true, capture: true })

    return () => {
      document.removeEventListener('mouseover', handleMouseOver)
      document.removeEventListener('mouseout', handleMouseOut)
      window.removeEventListener('scroll', handleDismiss, true)
      window.removeEventListener('click', handleDismiss, true)
      if (timerRef.current) {
        window.clearTimeout(timerRef.current)
      }
    }
  }, [data])

  useLayoutEffect(() => {
    if (!data) return

    const { rect, position } = data
    let pos = position

    // Flip position if near viewport edge
    if (pos === 'top' && rect.top < 45) {
      pos = 'bottom'
    } else if (pos === 'bottom' && rect.bottom > window.innerHeight - 45) {
      pos = 'top'
    } else if (pos === 'left' && rect.left < 70) {
      pos = 'right'
    } else if (pos === 'right' && rect.right > window.innerWidth - 70) {
      pos = 'left'
    }

    setActualPos(pos)

    let x = 0
    let y = 0

    if (pos === 'top') {
      x = rect.left + rect.width / 2
      y = rect.top - 6
    } else if (pos === 'bottom') {
      x = rect.left + rect.width / 2
      y = rect.bottom + 6
    } else if (pos === 'left') {
      x = rect.left - 6
      y = rect.top + rect.height / 2
    } else if (pos === 'right') {
      x = rect.right + 6
      y = rect.top + rect.height / 2
    }

    // Clamp horizontal position so tooltip doesn't bleed off screen edges
    if (tooltipRef.current) {
      const tooltipRect = tooltipRef.current.getBoundingClientRect()
      const halfWidth = tooltipRect.width / 2
      if (pos === 'top' || pos === 'bottom') {
        if (x - halfWidth < 8) {
          x = 8 + halfWidth
        } else if (x + halfWidth > window.innerWidth - 8) {
          x = window.innerWidth - 8 - halfWidth
        }
      }
    }

    setCoords({ x, y })
  }, [data])

  if (!data || typeof document === 'undefined') return null

  let transform = 'translate(-50%, -100%)'
  if (actualPos === 'bottom') {
    transform = 'translate(-50%, 0)'
  } else if (actualPos === 'left') {
    transform = 'translate(-100%, -50%)'
  } else if (actualPos === 'right') {
    transform = 'translate(0, -50%)'
  }

  return createPortal(
    <div
      ref={tooltipRef}
      className={`p-tooltip p-tooltip-visible p-tooltip-${actualPos}`}
      style={{
        left: `${coords.x}px`,
        top: `${coords.y}px`,
        transform
      }}
      role="tooltip"
    >
      <div className="p-tooltip-arrow" />
      <div className="p-tooltip-text">{data.text}</div>
    </div>,
    document.body
  )
}
