import { useState, useEffect } from 'react'
import { getSystems } from '../api'
import type { System } from '../types'

export function useSystem() {
  const [systems, setSystems] = useState<System[]>([])
  const [selected, setSelected] = useState<string>('')

  useEffect(() => {
    getSystems().then((data: System[]) => {
      setSystems(data ?? [])
      if (data?.length > 0) setSelected(data[0].code)
    }).catch(() => {})
  }, [])

  return { systems, selected, setSelected }
}
