import { Bike, Footprints, Waves, Snowflake, Activity, type LucideProps } from 'lucide-react'
import { getSportCategory, type SportCategory } from './sport-types'

const sportIcons: Record<SportCategory, typeof Activity> = {
  ride: Bike,
  run: Footprints,
  walk: Footprints,
  swim: Waves,
  winter: Snowflake,
  other: Activity,
}

/** SportIcon component that renders the appropriate icon for a sport type */
export function SportIcon({ sportType, ...props }: { sportType: string } & LucideProps) {
  const category = getSportCategory(sportType)
  const Icon = sportIcons[category]
  return <Icon {...props} />
}
