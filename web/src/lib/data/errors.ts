export class UnsupportedFeatureError extends Error {
  readonly feature: string

  constructor(feature: string, message?: string) {
    super(message ?? `${feature} is not supported in this mode`)
    this.name = 'UnsupportedFeatureError'
    this.feature = feature
  }
}

export function isUnsupportedFeatureError(err: unknown): err is UnsupportedFeatureError {
  return err instanceof UnsupportedFeatureError
}
