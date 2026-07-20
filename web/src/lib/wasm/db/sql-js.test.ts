import { afterEach, describe, expect, it, vi } from 'vitest'
import { DATABASE_LOCK_NAME, WasmDatabase } from './sql-js'

const originalNavigatorDescriptor = Object.getOwnPropertyDescriptor(globalThis, 'navigator')

afterEach(() => {
  if (originalNavigatorDescriptor) {
    Object.defineProperty(globalThis, 'navigator', originalNavigatorDescriptor)
  } else {
    Reflect.deleteProperty(globalThis, 'navigator')
  }
})

describe('WasmDatabase persistence locking', () => {
  it('holds the database Web Lock while persisting', async () => {
    let insideLock = false
    const save = vi.fn(async () => {
      expect(insideLock).toBe(true)
    })
    const request = vi.fn(
      async (_name: string, callback: () => Promise<unknown>): Promise<unknown> => {
        insideLock = true
        try {
          return await callback()
        } finally {
          insideLock = false
        }
      }
    )
    Object.defineProperty(globalThis, 'navigator', {
      configurable: true,
      value: { locks: { request } },
    })

    const database = new WasmDatabase({ autoPersistInterval: 0 })
    Object.assign(database, {
      db: { export: () => new Uint8Array([1, 2, 3]) },
      storage: { save },
    })

    await database.persist()

    expect(request).toHaveBeenCalledWith(DATABASE_LOCK_NAME, expect.any(Function))
    expect(save).toHaveBeenCalledOnce()
  })
})
