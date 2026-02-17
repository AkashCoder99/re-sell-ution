/**
 * F13 — Upload photos UI: multi-file, progress, reorder/delete, retry
 */

import { useState, useCallback } from 'react'

export interface PhotoItem {
  id: string
  file?: File
  preview: string
  url?: string
  progress: number
  error?: string
  status: 'pending' | 'uploading' | 'done' | 'failed'
}

interface PhotoUploadProps {
  maxFiles?: number
  maxSizeBytes?: number
  value: PhotoItem[]
  onChange: (items: PhotoItem[]) => void
  onUpload?: (file: File) => Promise<string>
  disabled?: boolean
}

const DEFAULT_MAX_FILES = 10
const DEFAULT_MAX_BYTES = 5 * 1024 * 1024 // 5MB

export default function PhotoUpload({
  maxFiles = DEFAULT_MAX_FILES,
  maxSizeBytes = DEFAULT_MAX_BYTES,
  value,
  onChange,
  onUpload,
  disabled = false
}: PhotoUploadProps) {
  const [dragOver, setDragOver] = useState(false)

  const addFiles = useCallback(
    (files: FileList | File[]) => {
      const list = Array.isArray(files) ? files : Array.from(files)
      const existing = value.length
      const toAdd = list.slice(0, Math.max(0, maxFiles - existing))
      const newItems: PhotoItem[] = toAdd.map((file) => {
        if (file.size > maxSizeBytes) {
          return {
            id: `err_${file.name}_${Date.now()}`,
            file,
            preview: '',
            progress: 0,
            error: 'File too large (max 5MB)',
            status: 'failed'
          }
        }
        return {
          id: `preview_${file.name}_${Date.now()}`,
          file,
          preview: URL.createObjectURL(file),
          progress: 0,
          status: 'pending' as const
        }
      })
      onChange([...value, ...newItems])
    },
    [value, maxFiles, maxSizeBytes, onChange]
  )

  const remove = useCallback(
    (id: string) => {
      const item = value.find((i) => i.id === id)
      if (item?.preview && item.preview.startsWith('blob:')) {
        URL.revokeObjectURL(item.preview)
      }
      onChange(value.filter((i) => i.id !== id))
    },
    [value, onChange]
  )

  const move = useCallback(
    (id: string, direction: 'left' | 'right') => {
      const idx = value.findIndex((i) => i.id === id)
      if (idx === -1) return
      const next = idx + (direction === 'right' ? 1 : -1)
      if (next < 0 || next >= value.length) return
      const nextList = [...value]
      ;[nextList[idx], nextList[next]] = [nextList[next], nextList[idx]]
      onChange(nextList)
    },
    [value, onChange]
  )

  const retry = useCallback(
    async (id: string) => {
      const item = value.find((i) => i.id === id)
      if (!item?.file || !onUpload) return
      const updated = value.map((i) =>
        i.id === id ? { ...i, status: 'uploading' as const, progress: 0, error: undefined } : i
      )
      onChange(updated)
      try {
        const url = await onUpload(item.file)
        onChange(
          value.map((i) =>
            i.id === id
              ? { ...i, url, progress: 100, status: 'done' as const, error: undefined }
              : i
          )
        )
      } catch (err) {
        onChange(
          value.map((i) =>
            i.id === id
              ? {
                  ...i,
                  status: 'failed' as const,
                  error: err instanceof Error ? err.message : 'Upload failed'
                }
              : i
          )
        )
      }
    },
    [value, onChange, onUpload]
  )

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files?.length) addFiles(files)
    e.target.value = ''
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    if (disabled) return
    const files = e.dataTransfer.files
    if (files?.length) addFiles(files)
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(true)
  }

  const handleDragLeave = () => setDragOver(false)

  return (
    <div className="photo-upload">
      <div
        className={`photo-upload-dropzone ${dragOver ? 'drag-over' : ''} ${disabled ? 'disabled' : ''}`}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
      >
        <input
          type="file"
          accept="image/*"
          multiple
          onChange={handleInputChange}
          disabled={disabled || value.length >= maxFiles}
          className="photo-upload-input"
        />
        <p className="photo-upload-hint">
          {value.length >= maxFiles
            ? `Maximum ${maxFiles} photos`
            : 'Drag & drop images here or click to browse'}
        </p>
        <p className="photo-upload-limit">Max 5MB per image</p>
      </div>

      {value.length > 0 && (
        <ul className="photo-upload-list">
          {value.map((item, index) => (
            <li key={item.id} className="photo-upload-item">
              <div className="photo-upload-preview-wrap">
                {item.preview ? (
                  <img
                    src={item.preview}
                    alt={`Upload ${index + 1}`}
                    className="photo-upload-preview"
                  />
                ) : (
                  <div className="photo-upload-placeholder">No preview</div>
                )}
                {item.status === 'uploading' && (
                  <div className="photo-upload-progress-wrap">
                    <div
                      className="photo-upload-progress-bar"
                      style={{ width: `${item.progress}%` }}
                    />
                  </div>
                )}
                {item.error && (
                  <p className="photo-upload-error-text">{item.error}</p>
                )}
              </div>
              <div className="photo-upload-actions">
                <button
                  type="button"
                  className="photo-upload-btn"
                  onClick={() => move(item.id, 'left')}
                  disabled={disabled || index === 0}
                  title="Move left"
                >
                  ←
                </button>
                <button
                  type="button"
                  className="photo-upload-btn"
                  onClick={() => move(item.id, 'right')}
                  disabled={disabled || index === value.length - 1}
                  title="Move right"
                >
                  →
                </button>
                {item.status === 'failed' && onUpload && (
                  <button
                    type="button"
                    className="photo-upload-btn retry"
                    onClick={() => retry(item.id)}
                    title="Retry"
                  >
                    Retry
                  </button>
                )}
                <button
                  type="button"
                  className="photo-upload-btn delete"
                  onClick={() => remove(item.id)}
                  disabled={disabled}
                  title="Remove"
                >
                  ✕
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
