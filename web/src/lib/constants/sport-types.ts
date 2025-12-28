export const SPORT_TYPES = {
  // Cycling
  Ride: { label: 'Ride', color: 'sport-ride', icon: 'bike' },
  MountainBikeRide: { label: 'Mountain Bike', color: 'sport-ride', icon: 'bike' },
  GravelRide: { label: 'Gravel Ride', color: 'sport-ride', icon: 'bike' },
  EBikeRide: { label: 'E-Bike Ride', color: 'sport-ride', icon: 'bike' },
  EMountainBikeRide: { label: 'E-Mountain Bike', color: 'sport-ride', icon: 'bike' },
  VirtualRide: { label: 'Virtual Ride', color: 'sport-ride', icon: 'bike' },
  Velomobile: { label: 'Velomobile', color: 'sport-ride', icon: 'bike' },

  // Running
  Run: { label: 'Run', color: 'sport-run', icon: 'footprints' },
  TrailRun: { label: 'Trail Run', color: 'sport-run', icon: 'footprints' },
  VirtualRun: { label: 'Virtual Run', color: 'sport-run', icon: 'footprints' },

  // Walking
  Walk: { label: 'Walk', color: 'sport-walk', icon: 'footprints' },
  Hike: { label: 'Hike', color: 'sport-walk', icon: 'mountain' },

  // Water Sports
  Swim: { label: 'Swim', color: 'sport-swim', icon: 'waves' },
  Canoeing: { label: 'Canoeing', color: 'sport-swim', icon: 'waves' },
  Kayaking: { label: 'Kayaking', color: 'sport-swim', icon: 'waves' },
  Rowing: { label: 'Rowing', color: 'sport-swim', icon: 'waves' },
  StandUpPaddling: { label: 'SUP', color: 'sport-swim', icon: 'waves' },
  Surfing: { label: 'Surfing', color: 'sport-swim', icon: 'waves' },
  Kitesurf: { label: 'Kitesurf', color: 'sport-swim', icon: 'wind' },
  Windsurf: { label: 'Windsurf', color: 'sport-swim', icon: 'wind' },

  // Winter Sports
  AlpineSki: { label: 'Alpine Ski', color: 'sport-winter', icon: 'ski' },
  BackcountrySki: { label: 'Backcountry Ski', color: 'sport-winter', icon: 'ski' },
  NordicSki: { label: 'Nordic Ski', color: 'sport-winter', icon: 'ski' },
  Snowboard: { label: 'Snowboard', color: 'sport-winter', icon: 'snowflake' },
  Snowshoe: { label: 'Snowshoe', color: 'sport-winter', icon: 'snowflake' },
  IceSkate: { label: 'Ice Skate', color: 'sport-winter', icon: 'snowflake' },

  // Skating
  InlineSkate: { label: 'Inline Skate', color: 'sport-other', icon: 'circle' },
  RollerSki: { label: 'Roller Ski', color: 'sport-other', icon: 'circle' },
  Skateboard: { label: 'Skateboard', color: 'sport-other', icon: 'circle' },

  // Fitness
  Crossfit: { label: 'Crossfit', color: 'sport-other', icon: 'dumbbell' },
  WeightTraining: { label: 'Weight Training', color: 'sport-other', icon: 'dumbbell' },
  Workout: { label: 'Workout', color: 'sport-other', icon: 'dumbbell' },
  Elliptical: { label: 'Elliptical', color: 'sport-other', icon: 'dumbbell' },
  StairStepper: { label: 'Stair Stepper', color: 'sport-other', icon: 'dumbbell' },
  VirtualRow: { label: 'Virtual Row', color: 'sport-other', icon: 'dumbbell' },
  HighIntensityIntervalTraining: { label: 'HIIT', color: 'sport-other', icon: 'dumbbell' },

  // Mind & Body
  Yoga: { label: 'Yoga', color: 'sport-other', icon: 'heart' },
  Pilates: { label: 'Pilates', color: 'sport-other', icon: 'heart' },

  // Racquet Sports
  Tennis: { label: 'Tennis', color: 'sport-other', icon: 'circle' },
  Badminton: { label: 'Badminton', color: 'sport-other', icon: 'circle' },
  Pickleball: { label: 'Pickleball', color: 'sport-other', icon: 'circle' },
  Racquetball: { label: 'Racquetball', color: 'sport-other', icon: 'circle' },
  Squash: { label: 'Squash', color: 'sport-other', icon: 'circle' },
  TableTennis: { label: 'Table Tennis', color: 'sport-other', icon: 'circle' },

  // Outdoor
  Golf: { label: 'Golf', color: 'sport-other', icon: 'flag' },
  RockClimbing: { label: 'Rock Climbing', color: 'sport-other', icon: 'mountain' },
  Sail: { label: 'Sailing', color: 'sport-other', icon: 'wind' },
  Soccer: { label: 'Soccer', color: 'sport-other', icon: 'circle' },

  // Adaptive
  Handcycle: { label: 'Handcycle', color: 'sport-ride', icon: 'bike' },
  Wheelchair: { label: 'Wheelchair', color: 'sport-other', icon: 'circle' },
} as const

export type SportType = keyof typeof SPORT_TYPES

export function getSportTypeInfo(sportType: string) {
  return (
    SPORT_TYPES[sportType as SportType] ?? {
      label: sportType,
      color: 'sport-other',
      icon: 'circle',
    }
  )
}

export const CYCLING_TYPES: SportType[] = [
  'Ride',
  'MountainBikeRide',
  'GravelRide',
  'EBikeRide',
  'EMountainBikeRide',
  'VirtualRide',
  'Velomobile',
  'Handcycle',
]

export const RUNNING_TYPES: SportType[] = ['Run', 'TrailRun', 'VirtualRun']
export const WALKING_TYPES: SportType[] = ['Walk', 'Hike']
export const WATER_TYPES: SportType[] = [
  'Swim',
  'Canoeing',
  'Kayaking',
  'Rowing',
  'StandUpPaddling',
  'Surfing',
  'Kitesurf',
  'Windsurf',
]
export const WINTER_TYPES: SportType[] = [
  'AlpineSki',
  'BackcountrySki',
  'NordicSki',
  'Snowboard',
  'Snowshoe',
  'IceSkate',
]
