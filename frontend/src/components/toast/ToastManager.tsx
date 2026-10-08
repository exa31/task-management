import { Toast } from 'primereact/toast'
import React, { useEffect, useRef } from 'react'

import { appToast } from './toastService'

export const AppToastContainer: React.FC = () => {
  const toastRef = useRef<Toast>(null)

  useEffect(() => {
    appToast.setInstance(toastRef.current)
    return () => {
      appToast.setInstance(null)
    }
  }, [])

  return <Toast ref={toastRef} position="top-right" baseZIndex={99999} className="z-[99999]" />
}
