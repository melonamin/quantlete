/**
 * OPFS (Origin Private File System) storage layer for sql.js database persistence.
 *
 * OPFS provides a high-performance file system API that's isolated per-origin,
 * making it ideal for storing the SQLite database in the browser.
 */

const DB_FILE_NAME = 'quantlete.db'
const DB_DIRECTORY = 'quantlete-data'

export interface OPFSStorageOptions {
  fileName?: string
  directoryName?: string
}

export class OPFSStorage {
  private fileName: string
  private directoryName: string
  private directoryHandle: FileSystemDirectoryHandle | null = null

  constructor(options: OPFSStorageOptions = {}) {
    this.fileName = options.fileName || DB_FILE_NAME
    this.directoryName = options.directoryName || DB_DIRECTORY
  }

  /**
   * Check if OPFS is supported in the current browser.
   */
  static isSupported(): boolean {
    return (
      typeof navigator !== 'undefined' &&
      'storage' in navigator &&
      'getDirectory' in navigator.storage
    )
  }

  /**
   * Initialize the OPFS directory handle.
   */
  private async getDirectory(): Promise<FileSystemDirectoryHandle> {
    if (this.directoryHandle) {
      return this.directoryHandle
    }

    if (!OPFSStorage.isSupported()) {
      throw new Error('OPFS is not supported in this browser')
    }

    const root = await navigator.storage.getDirectory()
    this.directoryHandle = await root.getDirectoryHandle(this.directoryName, {
      create: true,
    })

    return this.directoryHandle
  }

  /**
   * Load the database from OPFS.
   * Returns null if no database exists yet.
   */
  async load(): Promise<Uint8Array | null> {
    try {
      const dir = await this.getDirectory()
      const fileHandle = await dir.getFileHandle(this.fileName)
      const file = await fileHandle.getFile()
      const buffer = await file.arrayBuffer()

      if (buffer.byteLength === 0) {
        return null
      }

      console.log(`[OPFS] Loaded database: ${buffer.byteLength} bytes`)
      return new Uint8Array(buffer)
    } catch (error) {
      if ((error as Error).name === 'NotFoundError') {
        console.log('[OPFS] No existing database found')
        return null
      }
      throw error
    }
  }

  /**
   * Save the database to OPFS.
   */
  async save(data: Uint8Array): Promise<void> {
    const dir = await this.getDirectory()
    const fileHandle = await dir.getFileHandle(this.fileName, { create: true })

    // Use createWritable for atomic writes
    const writable = await fileHandle.createWritable()
    try {
      // Copy data to new ArrayBuffer for TypeScript compatibility
      // sql.js types use ArrayBufferLike which may include SharedArrayBuffer,
      // but in practice it's always a regular ArrayBuffer
      const buffer = new ArrayBuffer(data.byteLength)
      new Uint8Array(buffer).set(data)
      await writable.write(buffer)
      await writable.close()
      console.log(`[OPFS] Saved database: ${data.byteLength} bytes`)
    } catch (error) {
      await writable.abort()
      throw error
    }
  }

  /**
   * Delete the database file from OPFS.
   */
  async delete(): Promise<void> {
    try {
      const dir = await this.getDirectory()
      await dir.removeEntry(this.fileName)
      console.log('[OPFS] Database deleted')
    } catch (error) {
      if ((error as Error).name !== 'NotFoundError') {
        throw error
      }
    }
  }

  /**
   * Check if a database file exists.
   */
  async exists(): Promise<boolean> {
    try {
      const dir = await this.getDirectory()
      await dir.getFileHandle(this.fileName)
      return true
    } catch {
      return false
    }
  }

  /**
   * Get database file size in bytes.
   */
  async getSize(): Promise<number> {
    try {
      const dir = await this.getDirectory()
      const fileHandle = await dir.getFileHandle(this.fileName)
      const file = await fileHandle.getFile()
      return file.size
    } catch {
      return 0
    }
  }
}

/**
 * Fallback storage using IndexedDB for browsers without OPFS support.
 * This provides similar persistence but with lower performance.
 */
export class IndexedDBStorage {
  private dbName: string
  private storeName = 'database'

  constructor(dbName = 'quantlete-storage') {
    this.dbName = dbName
  }

  private openDB(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(this.dbName, 1)

      request.onerror = () => reject(request.error)
      request.onsuccess = () => resolve(request.result)

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result
        if (!db.objectStoreNames.contains(this.storeName)) {
          db.createObjectStore(this.storeName)
        }
      }
    })
  }

  async load(): Promise<Uint8Array | null> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const tx = db.transaction(this.storeName, 'readonly')
      const store = tx.objectStore(this.storeName)
      const request = store.get('db')

      request.onerror = () => reject(request.error)
      request.onsuccess = () => {
        const result = request.result as Uint8Array | undefined
        resolve(result || null)
      }
    })
  }

  async save(data: Uint8Array): Promise<void> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const tx = db.transaction(this.storeName, 'readwrite')
      const store = tx.objectStore(this.storeName)
      const request = store.put(data, 'db')

      request.onerror = () => reject(request.error)
      request.onsuccess = () => resolve()
    })
  }

  async delete(): Promise<void> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const tx = db.transaction(this.storeName, 'readwrite')
      const store = tx.objectStore(this.storeName)
      const request = store.delete('db')

      request.onerror = () => reject(request.error)
      request.onsuccess = () => resolve()
    })
  }

  async exists(): Promise<boolean> {
    const data = await this.load()
    return data !== null
  }

  async getSize(): Promise<number> {
    const data = await this.load()
    return data?.byteLength || 0
  }
}

/**
 * Get the best available storage backend.
 * Prefers OPFS but falls back to IndexedDB.
 */
export function getStorageBackend(): OPFSStorage | IndexedDBStorage {
  if (OPFSStorage.isSupported()) {
    console.log('[Storage] Using OPFS backend')
    return new OPFSStorage()
  }

  console.log('[Storage] OPFS not supported, falling back to IndexedDB')
  return new IndexedDBStorage()
}
