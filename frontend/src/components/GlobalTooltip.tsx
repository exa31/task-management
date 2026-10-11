import React, { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

type TooltipPosition = 'top' | 'bottom' | 'left' | 'right'

interface TooltipData {
  text: string
  position: TooltipPosition
  rect: DOMRect
  targetCenterX: number
}

/**
 * Universal Event-Delegated Tooltip Component.
 * Automatically displays sleek floating tooltips for ANY static or dynamically mounted DOM element
 * possessing `data-pr-tooltip` or `title` attributes across modals, drawers, tables, and standalone icons.
 */
export const GlobalTooltip: React.FC = () => {
  const [data, setData] = useState<TooltipData | null>(null)
  const [coords, setCoords] = useState<{ x: number; y: number }>({ x: 0, y: 0 })
  const [actualPos, setActualPos] = useState<TooltipPosition>('top')
  const [arrowShift, setArrowShift] = useState(0)
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
        let pos = (target.getAttribute('data-pr-position') || 'top') as TooltipPosition

        // Auto flip if near viewport edges
        if (pos === 'top' && rect.top < 40) {
          pos = 'bottom'
        } else if (pos === 'bottom' && rect.bottom > window.innerHeight - 40) {
          pos = 'top'
        } else if (pos === 'left' && rect.left < 80) {
          pos = 'right'
        } else if (pos === 'right' && rect.right > window.innerWidth - 80) {
          pos = 'left'
        }

        const targetCenterX = rect.left + rect.width / 2

        let initX = 0
        let initY = 0

        if (pos === 'top') {
          initX = targetCenterX
          initY = rect.top - 6
        } else if (pos === 'bottom') {
          initX = targetCenterX
          initY = rect.bottom + 6
        } else if (pos === 'left') {
          initX = rect.left - 6
          initY = rect.top + rect.height / 2
        } else if (pos === 'right') {
          initX = rect.right + 6
          initY = rect.top + rect.height / 2
        }

        setActualPos(pos)
        setCoords({ x: initX, y: initY })
        setArrowShift(0)
        setData({
          text: text.trim(),
          position: pos,
          rect,
          targetCenterX
        })
      }, 100)
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
    if (!data || !tooltipRef.current) return

    const { position, targetCenterX } = data
    const tooltipRect = tooltipRef.current.getBoundingClientRect()
    const tooltipWidth = tooltipRect.width
    const halfWidth = tooltipWidth / 2

    if (position === 'top' || position === 'bottom') {
      const minX = 12 + halfWidth
      const maxX = window.innerWidth - 12 - halfWidth

      let finalX = targetCenterX
      if (finalX < minX) {
        finalX = minX
      } else if (finalX > maxX) {
        finalX = maxX
      }

      // Calculate shift for arrow so it stays centered on target
      let shift = targetCenterX - finalX
      const maxShift = Math.max(0, halfWidth - 10)
      shift = Math.max(-maxShift, Math.min(maxShift, shift))

      setCoords((prev) => (prev.x === finalX ? prev : { ...prev, x: finalX }))
      setArrowShift(shift)
    }
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
      <div
        className="p-tooltip-arrow"
        style={arrowShift ? { transform: `translateX(calc(-50% + ${arrowShift}px))` } : undefined}
      />
      <div className="p-tooltip-text">{data.text}</div>
    </div>,
    document.body
  )
}
