import { useSettingsStore, type UnitSystem } from '@/stores/settings'
import {
  formatDistance,
  formatElevation,
  formatSpeed,
  formatPace,
  getDistanceUnit,
  getElevationUnit,
  getSpeedUnit,
  getPaceUnit,
} from '@/lib/format'

interface FormattedMetrics {
  unitSystem: UnitSystem
  formatDistance: (meters: number) => string
  formatElevation: (meters: number) => string
  formatSpeed: (metersPerSecond: number) => string
  formatPace: (metersPerSecond: number) => string
  distanceUnit: string
  elevationUnit: string
  speedUnit: string
  paceUnit: string
}

/**
 * Hook that provides unit-aware formatting functions.
 * Automatically uses the user's preferred unit system from settings.
 *
 * @example
 * const { formatDistance, formatElevation, distanceUnit } = useFormattedMetrics()
 * // formatDistance(5000) => "5.00 km" or "3.11 mi"
 */
export function useFormattedMetrics(): FormattedMetrics {
  const unitSystem = useSettingsStore((s) => s.unitSystem)

  return {
    unitSystem,
    formatDistance: (meters: number) => formatDistance(meters, unitSystem),
    formatElevation: (meters: number) => formatElevation(meters, unitSystem),
    formatSpeed: (metersPerSecond: number) => formatSpeed(metersPerSecond, unitSystem),
    formatPace: (metersPerSecond: number) => formatPace(metersPerSecond, unitSystem),
    distanceUnit: getDistanceUnit(unitSystem),
    elevationUnit: getElevationUnit(unitSystem),
    speedUnit: getSpeedUnit(unitSystem),
    paceUnit: getPaceUnit(unitSystem),
  }
}
