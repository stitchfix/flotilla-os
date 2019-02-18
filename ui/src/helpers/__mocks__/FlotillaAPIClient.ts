const getTasks = jest.fn()
const getTask = jest.fn()
const getTaskByAlias = jest.fn()
const getTaskHistory = jest.fn()
const createTask = jest.fn()
const updateTask = jest.fn()
const deleteTask = jest.fn()
const runTask = jest.fn()
const getActiveRuns = jest.fn()
const stopRun = jest.fn()
const getRun = jest.fn()
const getRunLogs = jest.fn()
const getGroups = jest.fn()
const getClusters = jest.fn()
const getTags = jest.fn()
const request = jest.fn()
const constructURL = jest.fn()
const processError = jest.fn()

export default jest.fn().mockImplementation(() => {
  return {
    getTasks,
    getTask,
    getTaskByAlias,
    getTaskHistory,
    createTask,
    updateTask,
    deleteTask,
    runTask,
    getActiveRuns,
    stopRun,
    getRun,
    getRunLogs,
    getGroups,
    getClusters,
    getTags,
    request,
    constructURL,
    processError,
  }
})
