const calculateDuration = (
  start: string,
  end: string | null | undefined
): number => {
  const s = Date.parse(start)
  const e = end ? Date.parse(end) : Date.now()

  if (isNaN(s) || isNaN(e)) return 0
  return e - s
}

export default calculateDuration
