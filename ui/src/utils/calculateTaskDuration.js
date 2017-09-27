import moment from 'moment'

function addZeroPadding(str) {
  const l = str.toString().length
  if (l === 1) return `0${str}`
  else return str
}

function formatDuration({ days, hours, minutes, seconds }) {
  let duration = ``
  if (days) {
    duration += `${days}.`
  }
  if (hours) {
    duration += `${addZeroPadding(hours)}:`
  } else {
    duration += `00:`
  }
  if (minutes) {
    duration += `${addZeroPadding(minutes)}:`
  } else {
    duration += `00:`
  }
  duration += `${addZeroPadding(seconds)}`
  return duration
}

export default function calculateTaskDuration(task) {
  let duration
  if (task.started_at) {
    const start = new Date(task.started_at)
    const end = !!task.finished_at ? new Date(task.finished_at) : new Date()
    const diff = (end - start)
    duration = formatDuration(moment.duration(diff)._data)
  } else {
    duration = '-'
  }
  return duration
}
