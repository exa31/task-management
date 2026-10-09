import React, { useEffect, useRef, useState } from 'react'
import * as THREE from 'three'

interface InteractiveHeroSceneProps {
  isDarkMode?: boolean
  className?: string
}

type KanbanViewMode = 'flow' | 'focus' | 'monolith'

export const InteractiveHeroScene: React.FC<InteractiveHeroSceneProps> = ({
  isDarkMode = true,
  className = '',
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const [activeMode, setActiveMode] = useState<KanbanViewMode>('flow')
  const [isInteracting, setIsInteracting] = useState(false)
  const [tasksMovedCount, setTasksMovedCount] = useState(0)
  const [resetCount, setResetCount] = useState(0)
  const modeRef = useRef<KanbanViewMode>('flow')
  const resetCountRef = useRef(0)

  useEffect(() => {
    modeRef.current = activeMode
  }, [activeMode])

  useEffect(() => {
    resetCountRef.current = resetCount
  }, [resetCount])

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    // 1. Scene, Isometric Perspective Camera, Renderer Setup
    const width = container.clientWidth || 520
    const height = container.clientHeight || 480

    const scene = new THREE.Scene()
    // Perspective with isometric-leaning position
    const camera = new THREE.PerspectiveCamera(38, width / height, 0.1, 1000)
    camera.position.set(7.5, 6.2, 9.5)
    camera.lookAt(0, 0.5, 0)

    const renderer = new THREE.WebGLRenderer({
      alpha: true,
      antialias: true,
      powerPreference: 'high-performance',
    })
    renderer.setSize(width, height)
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.35
    container.appendChild(renderer.domElement)

    // 2. Dynamic Lighting
    const ambientLight = new THREE.AmbientLight(isDarkMode ? 0x1e1b4b : 0xe2e8f0, 2.2)
    scene.add(ambientLight)

    const keyLight = new THREE.PointLight(isDarkMode ? 0x6366f1 : 0x4f46e5, 5.0, 50)
    keyLight.position.set(8, 10, 8)
    scene.add(keyLight)

    const rimLight = new THREE.PointLight(isDarkMode ? 0x10b981 : 0x059669, 3.5, 40)
    rimLight.position.set(-8, 6, -4)
    scene.add(rimLight)

    const fillLight = new THREE.PointLight(isDarkMode ? 0x06b6d4 : 0x0284c7, 3.0, 40)
    fillLight.position.set(0, -5, 6)
    scene.add(fillLight)

    // 3. Main Isometric Stage Hierarchy
    const stageGroup = new THREE.Group()
    scene.add(stageGroup)

    // Stage rotation baseline for iconic isometric view
    stageGroup.rotation.y = -Math.PI / 6
    stageGroup.position.y = -0.4

    // =========================================================================
    // BASE PODIUM (Sharp Geometric Isometric Plinth)
    // =========================================================================
    const plinthGeo = new THREE.BoxGeometry(6.6, 0.18, 4.2)
    const plinthMat = new THREE.MeshPhysicalMaterial({
      color: isDarkMode ? 0x0f172a : 0xf1f5f9,
      emissive: isDarkMode ? 0x020617 : 0xe2e8f0,
      emissiveIntensity: 0.3,
      roughness: 0.2,
      metalness: 0.8,
      clearcoat: 0.5,
    })
    const plinth = new THREE.Mesh(plinthGeo, plinthMat)
    plinth.position.y = -1.6
    stageGroup.add(plinth)

    const plinthEdgesGeo = new THREE.EdgesGeometry(plinthGeo)
    const plinthEdgesMat = new THREE.LineBasicMaterial({
      color: isDarkMode ? 0x334155 : 0x94a3b8,
      linewidth: 1.5,
    })
    const plinthEdges = new THREE.LineSegments(plinthEdgesGeo, plinthEdgesMat)
    plinth.add(plinthEdges)

    // Laser Grid Lines on Plinth Surface
    const gridLinesGroup = new THREE.Group()
    const gridMat = new THREE.LineBasicMaterial({
      color: isDarkMode ? 0x1e293b : 0xcbd5e1,
      transparent: true,
      opacity: 0.7,
    })
    for (let x = -3; x <= 3; x += 1.0) {
      const g = new THREE.BufferGeometry().setFromPoints([
        new THREE.Vector3(x, -1.5, -2.0),
        new THREE.Vector3(x, -1.5, 2.0),
      ])
      gridLinesGroup.add(new THREE.Line(g, gridMat))
    }
    for (let z = -2; z <= 2; z += 1.0) {
      const g = new THREE.BufferGeometry().setFromPoints([
        new THREE.Vector3(-3.2, -1.5, z),
        new THREE.Vector3(3.2, -1.5, z),
      ])
      gridLinesGroup.add(new THREE.Line(g, gridMat))
    }
    stageGroup.add(gridLinesGroup)

    // =========================================================================
    // 3 KANBAN MONOLITHS (Columns: Backlog, In Progress, Done)
    // =========================================================================
    const columnConfigs = [
      {
        id: 'backlog',
        title: 'BACKLOG',
        x: -2.15,
        height: 2.8,
        color: isDarkMode ? 0x4338ca : 0x6366f1,
        emissive: isDarkMode ? 0x1e1b4b : 0x312e81,
        accent: 0x818cf8,
      },
      {
        id: 'in_progress',
        title: 'IN PROGRESS',
        x: 0.0,
        height: 3.4,
        color: isDarkMode ? 0x0284c7 : 0x0ea5e9,
        emissive: isDarkMode ? 0x082f49 : 0x0369a1,
        accent: 0x38bdf8,
      },
      {
        id: 'done',
        title: 'COMPLETED',
        x: 2.15,
        height: 3.0,
        color: isDarkMode ? 0x047857 : 0x10b981,
        emissive: isDarkMode ? 0x022c22 : 0x065f46,
        accent: 0x34d399,
      },
    ]

    const columnMeshes: THREE.Mesh[] = []

    columnConfigs.forEach((cfg) => {
      const colGeo = new THREE.BoxGeometry(1.65, cfg.height, 1.25)
      const colMat = new THREE.MeshPhysicalMaterial({
        color: cfg.color,
        emissive: cfg.emissive,
        emissiveIntensity: 0.4,
        roughness: 0.15,
        metalness: 0.3,
        transmission: 0.8, // Acrylic glass appearance
        ior: 1.52,
        transparent: true,
        opacity: 0.88,
      })
      const colMesh = new THREE.Mesh(colGeo, colMat)
      colMesh.position.set(cfg.x, -1.5 + cfg.height / 2, 0)
      stageGroup.add(colMesh)
      columnMeshes.push(colMesh)

      // Sharp Neon Edges
      const edgesGeo = new THREE.EdgesGeometry(colGeo)
      const edgesMat = new THREE.LineBasicMaterial({
        color: cfg.accent,
        linewidth: 2,
      })
      const edges = new THREE.LineSegments(edgesGeo, edgesMat)
      colMesh.add(edges)

      // Top Cap Luminous Neon Frame
      const topCapGeo = new THREE.BoxGeometry(1.66, 0.06, 1.26)
      const topCapMat = new THREE.MeshStandardMaterial({
        color: cfg.accent,
        emissive: cfg.accent,
        emissiveIntensity: 0.9,
        metalness: 0.8,
        roughness: 0.2,
      })
      const topCap = new THREE.Mesh(topCapGeo, topCapMat)
      topCap.position.y = cfg.height / 2 + 0.03
      colMesh.add(topCap)
    })

    // =========================================================================
    // FLOATING 3D TASK CARDS (Rectangular High-Tech Chips with Neon Borders)
    // =========================================================================
    const cardGroup = new THREE.Group()
    stageGroup.add(cardGroup)

    const cardGeo = new THREE.BoxGeometry(1.2, 0.65, 0.06)
    const cardEdgesGeo = new THREE.EdgesGeometry(cardGeo)

    interface TaskCard3D {
      mesh: THREE.Mesh
      baseX: number
      baseY: number
      baseZ: number
      targetX: number
      targetY: number
      targetZ: number
      isFlying?: boolean
      flightProgress?: number
      colIdx: number
    }

    const taskCards: TaskCard3D[] = []

    const cardDefinitions = [
      // Backlog Cards
      { colIdx: 0, x: -2.15, y: 0.8, z: 0.15, color: 0x6366f1, accent: 0xa5b4fc },
      { colIdx: 0, x: -2.15, y: 1.6, z: -0.1, color: 0x4f46e5, accent: 0x818cf8 },
      // In Progress Cards
      { colIdx: 1, x: 0.0, y: 1.2, z: 0.2, color: 0x0ea5e9, accent: 0x7dd3fc },
      { colIdx: 1, x: 0.0, y: 2.1, z: -0.15, color: 0x0284c7, accent: 0x38bdf8 },
      // The Flying / Transiting Task Card (Moving between columns)
      { colIdx: 1, x: 0.0, y: 2.8, z: 0.25, color: 0xec4899, accent: 0xf472b6, isFlying: true },
      // Done Cards
      { colIdx: 2, x: 2.15, y: 0.9, z: 0.15, color: 0x10b981, accent: 0x6ee7b7 },
      { colIdx: 2, x: 2.15, y: 1.7, z: -0.1, color: 0x059669, accent: 0x34d399 },
    ]

    cardDefinitions.forEach((def) => {
      const cardMat = new THREE.MeshPhysicalMaterial({
        color: def.color,
        emissive: def.color,
        emissiveIntensity: 0.45,
        metalness: 0.85,
        roughness: 0.15,
        clearcoat: 1.0,
      })
      const cardMesh = new THREE.Mesh(cardGeo, cardMat)
      cardMesh.position.set(def.x, def.y, def.z)
      // Slight pitch tilt for isometric readability
      cardMesh.rotation.x = -Math.PI / 12
      cardGroup.add(cardMesh)

      // Glowing Neon Rim
      const cardEdges = new THREE.LineSegments(
        cardEdgesGeo,
        new THREE.LineBasicMaterial({ color: def.accent, linewidth: 2 })
      )
      cardMesh.add(cardEdges)

      // Microchip Corner Tag
      const chipGeo = new THREE.BoxGeometry(0.25, 0.15, 0.08)
      const chipMat = new THREE.MeshBasicMaterial({ color: 0xffffff })
      const chip = new THREE.Mesh(chipGeo, chipMat)
      chip.position.set(-0.45, 0.2, 0.01)
      cardMesh.add(chip)

      taskCards.push({
        mesh: cardMesh,
        baseX: def.x,
        baseY: def.y,
        baseZ: def.z,
        targetX: def.x,
        targetY: def.y,
        targetZ: def.z,
        isFlying: def.isFlying || false,
        flightProgress: 0,
        colIdx: def.colIdx,
      })
    })

    // =========================================================================
    // LASER TRANSIT PIPELINE (Connecting the 3 Monoliths)
    // =========================================================================
    const transitBeamGeo = new THREE.BufferGeometry().setFromPoints([
      new THREE.Vector3(-2.15, 2.7, 0),
      new THREE.Vector3(0.0, 3.2, 0),
      new THREE.Vector3(2.15, 2.9, 0),
    ])
    const transitBeamMat = new THREE.LineBasicMaterial({
      color: 0x38bdf8,
      transparent: true,
      opacity: 0.85,
      linewidth: 2,
    })
    const transitBeam = new THREE.Line(transitBeamGeo, transitBeamMat)
    stageGroup.add(transitBeam)

    // Data Flow Particle Beacons
    const beaconCount = 40
    const beaconGeo = new THREE.BufferGeometry()
    const beaconPos = new Float32Array(beaconCount * 3)
    for (let i = 0; i < beaconCount; i++) {
      const t = Math.random()
      beaconPos[i * 3] = -2.15 + t * 4.3 // X across columns
      beaconPos[i * 3 + 1] = 2.7 + Math.sin(t * Math.PI) * 0.5 + (Math.random() - 0.5) * 0.2 // Y
      beaconPos[i * 3 + 2] = (Math.random() - 0.5) * 0.4 // Z
    }
    beaconGeo.setAttribute('position', new THREE.BufferAttribute(beaconPos, 3))
    const beaconMat = new THREE.PointsMaterial({
      color: 0x6ee7b7,
      size: 0.08,
      transparent: true,
      opacity: 0.85,
    })
    const beacons = new THREE.Points(beaconGeo, beaconMat)
    stageGroup.add(beacons)

    // =========================================================================
    // INTERACTION TRACKING & PERSISTENT ORBIT PHYSICS (DRAG ONLY)
    // =========================================================================
    let baseRotationY = -Math.PI / 6
    let baseRotationX = 0.08
    let targetPresetY = -Math.PI / 6
    let targetPresetX = 0.08
    let isTransitioningToPreset = false

    let isDragging = false
    let previousPointerX = 0
    let previousPointerY = 0
    let dragVelocityX = 0
    let dragVelocityY = 0
    let totalDragDistance = 0
    let hoverPulse = 1.0

    let previousMode: KanbanViewMode = modeRef.current
    let lastHandledReset = resetCountRef.current

    const handlePointerDown = (e: PointerEvent) => {
      if (e.button !== 0 && e.pointerType === 'mouse') return
      isDragging = true
      isTransitioningToPreset = false
      setIsInteracting(true)
      previousPointerX = e.clientX
      previousPointerY = e.clientY
      totalDragDistance = 0
      dragVelocityX = 0
      dragVelocityY = 0
      try {
        container.setPointerCapture(e.pointerId)
      } catch {
        // Fallback for browsers without pointer capture
      }
    }

    const handlePointerMove = (e: PointerEvent) => {
      // Strictly do nothing when mouse is moved without dragging
      if (!isDragging) return

      const deltaX = e.clientX - previousPointerX
      const deltaY = e.clientY - previousPointerY
      totalDragDistance += Math.hypot(deltaX, deltaY)

      dragVelocityY = deltaX * 0.007
      dragVelocityX = deltaY * 0.007

      baseRotationY += dragVelocityY
      baseRotationX = Math.max(-0.45, Math.min(0.65, baseRotationX + dragVelocityX))

      previousPointerX = e.clientX
      previousPointerY = e.clientY
    }

    const handlePointerUp = (e: PointerEvent) => {
      if (isDragging) {
        isDragging = false
        setIsInteracting(false)
        try {
          container.releasePointerCapture(e.pointerId)
        } catch {
          // Ignore
        }
      }
    }

    const handlePointerCancel = (e: PointerEvent) => {
      if (isDragging) {
        isDragging = false
        setIsInteracting(false)
        try {
          container.releasePointerCapture(e.pointerId)
        } catch {
          // Ignore
        }
      }
    }

    // Click triggers dynamic task movement animation (only if not dragged)
    const handleClick = () => {
      if (totalDragDistance > 6) return

      hoverPulse = 1.15
      setTasksMovedCount((prev) => prev + 1)

      // Trigger the flying card to cycle columns
      const flyingCard = taskCards.find((c) => c.isFlying)
      if (flyingCard) {
        flyingCard.flightProgress = 0
      }
    }

    container.addEventListener('pointerdown', handlePointerDown)
    container.addEventListener('pointermove', handlePointerMove)
    container.addEventListener('pointerup', handlePointerUp)
    container.addEventListener('pointercancel', handlePointerCancel)
    window.addEventListener('pointerup', handlePointerUp)
    container.addEventListener('click', handleClick)

    // Resize Observer
    const resizeObserver = new ResizeObserver(() => {
      if (!container) return
      const newW = container.clientWidth
      const newH = container.clientHeight
      camera.aspect = newW / newH
      camera.updateProjectionMatrix()
      renderer.setSize(newW, newH)
    })
    resizeObserver.observe(container)

    // 4. Render Animation Loop
    let animationFrameId: number
    const clock = new THREE.Clock()

    const animate = () => {
      animationFrameId = requestAnimationFrame(animate)
      clock.getDelta()
      const elapsedTime = clock.getElapsedTime()
      const curMode = modeRef.current

      // Handle Preset Transition when mode or reset changes
      if (curMode !== previousMode) {
        previousMode = curMode
        isTransitioningToPreset = true
        if (curMode === 'focus') {
          targetPresetY = 0.0
          targetPresetX = 0.12
        } else if (curMode === 'flow') {
          targetPresetY = -Math.PI / 4
          targetPresetX = 0.1
        } else {
          // monolith
          targetPresetY = -Math.PI / 6
          targetPresetX = 0.08
        }
      }

      if (resetCountRef.current !== lastHandledReset) {
        lastHandledReset = resetCountRef.current
        targetPresetY = -Math.PI / 6
        targetPresetX = 0.08
        isTransitioningToPreset = true
      }

      // Smooth preset animation
      if (isTransitioningToPreset && !isDragging) {
        baseRotationY += (targetPresetY - baseRotationY) * 0.08
        baseRotationX += (targetPresetX - baseRotationX) * 0.08
        if (Math.abs(targetPresetY - baseRotationY) < 0.003 && Math.abs(targetPresetX - baseRotationX) < 0.003) {
          baseRotationY = targetPresetY
          baseRotationX = targetPresetX
          isTransitioningToPreset = false
        }
      }

      // Smooth inertia when drag is released (stays permanently at user's angle!)
      if (!isDragging && !isTransitioningToPreset) {
        if (Math.abs(dragVelocityX) > 0.0001 || Math.abs(dragVelocityY) > 0.0001) {
          dragVelocityX *= 0.90
          dragVelocityY *= 0.90
          baseRotationY += dragVelocityY
          baseRotationX = Math.max(-0.45, Math.min(0.65, baseRotationX + dragVelocityX))
        } else {
          dragVelocityX = 0
          dragVelocityY = 0
        }
      }

      // Applied rotation: strictly locked to user's dragged angle (never wobbles on mouse move!)
      stageGroup.rotation.y = baseRotationY
      stageGroup.rotation.x = baseRotationX

      // Pulse recovery
      hoverPulse += (1.0 - hoverPulse) * 0.08
      stageGroup.scale.setScalar(hoverPulse)

      // =========================================================================
      // DYNAMIC BEHAVIOR PER MODE
      // =========================================================================
      taskCards.forEach((card, idx) => {
        if (card.isFlying) {
          // Card actively flying along the laser pipeline from In Progress to Done
          card.flightProgress = (card.flightProgress || 0) + 0.012
          if (card.flightProgress > 1) card.flightProgress = 0

          const p = card.flightProgress
          // Smooth arching trajectory
          card.mesh.position.x = 0.0 + p * 2.15
          card.mesh.position.y = 2.1 + Math.sin(p * Math.PI) * 1.1
          card.mesh.position.z = Math.sin(p * Math.PI * 2) * 0.35
          card.mesh.rotation.z = Math.sin(p * Math.PI) * 0.25
        } else {
          // Floating idle oscillation
          const floatOffset = Math.sin(elapsedTime * 2.0 + idx * 1.2) * 0.08
          card.mesh.position.y = card.baseY + floatOffset

          if (curMode === 'focus') {
            // Focus mode highlights column 1 (In Progress) and dims/drops others
            if (card.colIdx === 1) {
              card.mesh.scale.setScalar(1.15)
              card.mesh.position.z = card.baseZ + 0.35
            } else {
              card.mesh.scale.setScalar(0.9)
              card.mesh.position.z = card.baseZ - 0.2
            }
          } else {
            card.mesh.scale.setScalar(1.0)
            card.mesh.position.z = card.baseZ
          }
        }
      })

      // Column lighting oscillation in Focus Mode
      if (curMode === 'focus') {
        columnMeshes[1].scale.y = 1.05
        columnMeshes[0].scale.y = 0.95
        columnMeshes[2].scale.y = 0.95
      } else {
        columnMeshes.forEach((col) => {
          col.scale.y = 1.0
        })
      }

      // Animate laser beam glow & beacons
      const positions = beacons.geometry.attributes.position.array as Float32Array
      for (let i = 0; i < beaconCount; i++) {
        positions[i * 3] += 0.035
        if (positions[i * 3] > 2.15) {
          positions[i * 3] = -2.15
        }
      }
      beacons.geometry.attributes.position.needsUpdate = true

      // Floating ambient light drift
      keyLight.position.x = 8 + Math.sin(elapsedTime * 0.5) * 2
      rimLight.position.y = 6 + Math.cos(elapsedTime * 0.6) * 1.5

      renderer.render(scene, camera)
    }

    animate()

    // Cleanup
    return () => {
      cancelAnimationFrame(animationFrameId)
      container.removeEventListener('pointerdown', handlePointerDown)
      container.removeEventListener('pointermove', handlePointerMove)
      container.removeEventListener('pointerup', handlePointerUp)
      container.removeEventListener('pointercancel', handlePointerCancel)
      window.removeEventListener('pointerup', handlePointerUp)
      container.removeEventListener('click', handleClick)
      resizeObserver.disconnect()
      renderer.dispose()
      if (container.contains(renderer.domElement)) {
        container.removeChild(renderer.domElement)
      }
    }
  }, [isDarkMode])

  return (
    <div className={`relative flex flex-col items-center select-none ${className}`}>
      {/* 3D WebGL Canvas Container */}
      <div
        ref={containerRef}
        className="w-full h-[380px] sm:h-[470px] md:h-[510px] cursor-grab active:cursor-grabbing touch-none relative"
        title="Click to move tasks • Drag to rotate 3D Isometric Kanban"
      />

      {/* Floating Interactive Widget Bar */}
      <div className="absolute bottom-2 inset-x-0 flex items-center justify-center pointer-events-none px-2">
        <div className="pointer-events-auto flex flex-wrap items-center justify-center gap-2 px-3.5 py-2 rounded-xl bg-white/90 dark:bg-slate-900/90 backdrop-blur-2xl border border-slate-200/80 dark:border-slate-800/80 shadow-xl text-xs">
          <span className="text-[10px] uppercase font-mono tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
            <span
              className={`w-2 h-2 rounded-full ${
                isInteracting ? 'bg-emerald-400 animate-ping' : 'bg-indigo-500'
              }`}
            />
            <span className="font-bold text-slate-800 dark:text-slate-200">3D Kanban Monolith</span>
          </span>

          <div className="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 p-0.5 rounded-lg">
            <button
              onClick={() => setActiveMode('flow')}
              className={`px-2.5 py-1 rounded-md text-[11px] font-semibold transition-all cursor-pointer ${
                activeMode === 'flow'
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              Sprint Flow
            </button>
            <button
              onClick={() => setActiveMode('focus')}
              className={`px-2.5 py-1 rounded-md text-[11px] font-semibold transition-all cursor-pointer ${
                activeMode === 'focus'
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              Horizon Focus
            </button>
            <button
              onClick={() => setActiveMode('monolith')}
              className={`px-2.5 py-1 rounded-md text-[11px] font-semibold transition-all cursor-pointer ${
                activeMode === 'monolith'
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              All Columns
            </button>
            <button
              onClick={() => {
                setActiveMode('monolith')
                setResetCount((c) => c + 1)
              }}
              className="px-2 py-1 rounded-md text-[11px] font-medium text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 transition-colors flex items-center gap-1 cursor-pointer"
              title="Reset view to default isometric perspective"
            >
              <i className="pi pi-replay text-[9px]" />
              <span>Reset</span>
            </button>
          </div>

          <span className="text-[10px] text-slate-400 font-mono hidden sm:inline pl-1">
            (Click to move task • Drag 360°)
          </span>

          {tasksMovedCount > 0 && (
            <span className="text-[10px] bg-emerald-500/10 text-emerald-500 dark:text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded-full font-mono font-bold animate-fade-in">
              +{tasksMovedCount} Tasks Transited
            </span>
          )}
        </div>
      </div>
    </div>
  )
}

export default InteractiveHeroScene
