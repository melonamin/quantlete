/**
 * WASM Algorithms - Browser bindings for TinyGo-compiled algorithms.
 *
 * This module loads the shared algorithms.wasm (compiled with TinyGo) and provides
 * a typed interface for calling the WASM functions from the browser.
 *
 * Memory model: Uses fixed buffers for input/output data transfer.
 * - Copy input data to inputBuffer
 * - Call WASM function with length
 * - Read results from outputBuffer or intOutputBuffer
 */

// Types matching the TinyGo WASM exports
interface TinyGoExports {
  memory: WebAssembly.Memory

  // Buffer pointer getters
  getInputBufferPtr(): number
  getOutputBufferPtr(): number
  getIntOutputBufferPtr(): number

  // Power functions
  normalizedPower(length: number): number
  rollingMaxAverage(length: number, windowSeconds: number): number
  intensityFactor(np: number, ftp: number): number
  trainingStressScore(durationSeconds: number, np: number, ftp: number): number

  // Eddington functions
  eddingtonNumber(length: number): number
  eddingtonNextSteps(length: number, currentE: number, stepsToCalculate: number): number
  eddingtonHistory(length: number): number

  // Training load functions
  calculateTrainingLoad(length: number, ctlTau: number, atlTau: number): number
  calculateTrainingLoadWithInitial(
    length: number,
    initialCtl: number,
    initialAtl: number,
    ctlTau: number,
    atlTau: number
  ): number
  predictAfterWorkout(
    currentCtl: number,
    currentAtl: number,
    plannedTss: number,
    ctlTau: number,
    atlTau: number
  ): void
  tssForTargetTsb(
    currentCtl: number,
    currentAtl: number,
    targetTsb: number,
    ctlTau: number,
    atlTau: number
  ): number
}

let wasmExports: TinyGoExports | null = null
let initPromise: Promise<void> | null = null

// Cached buffer pointers
let inputPtr = 0
let outputPtr = 0
let intOutputPtr = 0

/**
 * Initialize the WASM module. Must be called before using any algorithm functions.
 */
export async function initAlgorithms(): Promise<void> {
  if (wasmExports) return
  if (initPromise) return initPromise

  initPromise = (async () => {
    // Load TinyGo's wasm_exec.js if not already loaded
    if (typeof (globalThis as unknown as { Go: unknown }).Go === 'undefined') {
      await loadWasmExec()
    }

    const response = await fetch('/wasm/algorithms.wasm')
    const bytes = await response.arrayBuffer()

    // Create Go instance (TinyGo runtime)
    const go = new (globalThis as unknown as { Go: new () => Go }).Go()
    const { instance } = await WebAssembly.instantiate(bytes, go.importObject)

    // Start the Go runtime (required for TinyGo)
    // This is fire-and-forget; TinyGo's main() returns immediately
    go.run(instance)

    wasmExports = instance.exports as unknown as TinyGoExports

    // Cache buffer pointers
    inputPtr = wasmExports.getInputBufferPtr()
    outputPtr = wasmExports.getOutputBufferPtr()
    intOutputPtr = wasmExports.getIntOutputBufferPtr()
  })()

  return initPromise
}

/**
 * Load TinyGo's wasm_exec.js dynamically.
 */
async function loadWasmExec(): Promise<void> {
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = '/wasm/wasm_exec.js'
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Failed to load wasm_exec.js'))
    document.head.appendChild(script)
  })
}

// TinyGo's Go type (from wasm_exec.js)
interface Go {
  importObject: WebAssembly.Imports
  run(instance: WebAssembly.Instance): Promise<void>
}

/**
 * Check if algorithms are initialized.
 */
export function isInitialized(): boolean {
  return wasmExports !== null
}

/**
 * Copy a number array to the WASM input buffer.
 */
function copyToInputBuffer(data: number[]): void {
  if (!wasmExports) throw new Error('WASM not initialized')
  const view = new Float64Array(wasmExports.memory.buffer, inputPtr, data.length)
  view.set(data)
}

/**
 * Read from the WASM output buffer (float64).
 */
function readOutputBuffer(length: number): number[] {
  if (!wasmExports) throw new Error('WASM not initialized')
  const view = new Float64Array(wasmExports.memory.buffer, outputPtr, length)
  return Array.from(view)
}

/**
 * Read from the WASM int output buffer (int32).
 */
function readIntOutputBuffer(length: number): number[] {
  if (!wasmExports) throw new Error('WASM not initialized')
  const view = new Int32Array(wasmExports.memory.buffer, intOutputPtr, length)
  return Array.from(view)
}

// ============================================================================
// Power Analysis
// ============================================================================

/**
 * Calculate Normalized Power from power data.
 *
 * @param watts - Power data (1 sample per second)
 * @returns Normalized power in watts
 */
export function normalizedPower(watts: number[]): number {
  if (!wasmExports) throw new Error('WASM not initialized')
  if (watts.length === 0) return 0

  copyToInputBuffer(watts)
  return wasmExports.normalizedPower(watts.length)
}

/**
 * Find the maximum rolling average for a given window size.
 *
 * @param values - Data array (1 sample per second)
 * @param windowSeconds - Window size in seconds
 * @returns Maximum average value over any window
 */
export function rollingMaxAverage(values: number[], windowSeconds: number): number {
  if (!wasmExports) throw new Error('WASM not initialized')
  if (values.length === 0) return 0

  copyToInputBuffer(values)
  return wasmExports.rollingMaxAverage(values.length, windowSeconds)
}

/**
 * Calculate Intensity Factor (NP / FTP).
 */
export function intensityFactor(np: number, ftp: number): number {
  if (!wasmExports) throw new Error('WASM not initialized')
  return wasmExports.intensityFactor(np, ftp)
}

/**
 * Calculate Training Stress Score.
 */
export function trainingStressScore(durationSeconds: number, np: number, ftp: number): number {
  if (!wasmExports) throw new Error('WASM not initialized')
  return wasmExports.trainingStressScore(durationSeconds, np, ftp)
}

// ============================================================================
// Eddington Number
// ============================================================================

/**
 * Calculate Eddington number from daily distances.
 *
 * @param distances - Array of daily distances in km
 * @returns Eddington number
 */
export function eddingtonNumber(distances: number[]): number {
  if (!wasmExports) throw new Error('WASM not initialized')
  if (distances.length === 0) return 0

  copyToInputBuffer(distances)
  return wasmExports.eddingtonNumber(distances.length)
}

/**
 * Calculate days needed for next Eddington numbers.
 *
 * @param distances - Array of daily distances in km
 * @param currentE - Current Eddington number
 * @param stepsToCalculate - How many steps to calculate
 * @returns Array of {target, daysNeeded} objects
 */
export function eddingtonNextSteps(
  distances: number[],
  currentE: number,
  stepsToCalculate: number = 5
): { target: number; daysNeeded: number }[] {
  if (!wasmExports) throw new Error('WASM not initialized')

  copyToInputBuffer(distances)
  const count = wasmExports.eddingtonNextSteps(distances.length, currentE, stepsToCalculate)
  const raw = readIntOutputBuffer(count * 2)

  const result: { target: number; daysNeeded: number }[] = []
  for (let i = 0; i < count; i++) {
    result.push({
      target: raw[i * 2],
      daysNeeded: raw[i * 2 + 1],
    })
  }
  return result
}

/**
 * Calculate progressive Eddington number over time.
 *
 * @param distances - Array of daily distances in km (chronological)
 * @returns Array of Eddington numbers (one per day)
 */
export function eddingtonHistory(distances: number[]): number[] {
  if (!wasmExports) throw new Error('WASM not initialized')
  if (distances.length === 0) return []

  copyToInputBuffer(distances)
  const count = wasmExports.eddingtonHistory(distances.length)
  return readIntOutputBuffer(count)
}

// ============================================================================
// Training Load
// ============================================================================

export interface TrainingLoadPoint {
  ctl: number
  atl: number
  tsb: number
}

/**
 * Calculate training load metrics from daily TSS values.
 *
 * @param dailyTss - Array of daily TSS values (chronological)
 * @param ctlTau - CTL time constant (default 42)
 * @param atlTau - ATL time constant (default 7)
 * @returns Array of {ctl, atl, tsb} for each day
 */
export function calculateTrainingLoad(
  dailyTss: number[],
  ctlTau: number = 42,
  atlTau: number = 7
): TrainingLoadPoint[] {
  if (!wasmExports) throw new Error('WASM not initialized')
  if (dailyTss.length === 0) return []

  copyToInputBuffer(dailyTss)
  const count = wasmExports.calculateTrainingLoad(dailyTss.length, ctlTau, atlTau)
  const raw = readOutputBuffer(count * 3)

  const result: TrainingLoadPoint[] = []
  for (let i = 0; i < count; i++) {
    result.push({
      ctl: raw[i * 3],
      atl: raw[i * 3 + 1],
      tsb: raw[i * 3 + 2],
    })
  }
  return result
}

/**
 * Calculate training load starting from existing CTL/ATL values.
 */
export function calculateTrainingLoadWithInitial(
  dailyTss: number[],
  initialCtl: number,
  initialAtl: number,
  ctlTau: number = 42,
  atlTau: number = 7
): TrainingLoadPoint[] {
  if (!wasmExports) throw new Error('WASM not initialized')
  if (dailyTss.length === 0) return []

  copyToInputBuffer(dailyTss)
  const count = wasmExports.calculateTrainingLoadWithInitial(
    dailyTss.length,
    initialCtl,
    initialAtl,
    ctlTau,
    atlTau
  )
  const raw = readOutputBuffer(count * 3)

  const result: TrainingLoadPoint[] = []
  for (let i = 0; i < count; i++) {
    result.push({
      ctl: raw[i * 3],
      atl: raw[i * 3 + 1],
      tsb: raw[i * 3 + 2],
    })
  }
  return result
}

/**
 * Calculate predicted CTL/ATL/TSB after a planned workout.
 */
export function predictAfterWorkout(
  currentCtl: number,
  currentAtl: number,
  plannedTss: number,
  ctlTau: number = 42,
  atlTau: number = 7
): TrainingLoadPoint {
  if (!wasmExports) throw new Error('WASM not initialized')

  wasmExports.predictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau)
  const raw = readOutputBuffer(3)

  return {
    ctl: raw[0],
    atl: raw[1],
    tsb: raw[2],
  }
}

/**
 * Calculate TSS needed to achieve a target TSB.
 */
export function tssForTargetTsb(
  currentCtl: number,
  currentAtl: number,
  targetTsb: number,
  ctlTau: number = 42,
  atlTau: number = 7
): number {
  if (!wasmExports) throw new Error('WASM not initialized')
  return wasmExports.tssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau)
}
