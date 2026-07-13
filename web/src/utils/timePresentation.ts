export function hasDisplayableTime(value?: string | null): boolean {
  if (!value) {
    return false
  }

  const date = new Date(value)
  return !Number.isNaN(date.getTime()) && date.getFullYear() > 1
}

export function formatDisplayableTime(value?: string | null): string {
  if (!hasDisplayableTime(value)) {
    return '未知'
  }

  return new Date(value!).toLocaleString()
}
