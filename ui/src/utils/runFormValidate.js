export default function runFormValidate(values) {
  const defaultErrMessage = "This is a required field."
  const errors = {}
  if (!values.cluster) errors.cluster = "You must select a cluster."

  if (!!values.env && values.env.length > 0) {
    const envErr = []
    values.env.forEach((e, i) => {
      const envvarErr = {}
      if (!!e) {
        if (!e.value) {
          envvarErr.value = defaultErrMessage
          envErr[i] = envvarErr
        }
        if (!e.name) {
          envvarErr.name = defaultErrMessage
          envErr[i] = envvarErr
        }
      }
    })
    if (envErr.length > 0) {
      errors.env = envErr
    }
  }
  return errors
}
