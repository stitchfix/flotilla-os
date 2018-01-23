import { invalidEnv } from '../../constants/'

export default function validate(values) {
  const requiredFields = [
    { name: 'alias', errorMessage: 'This is a required field.' },
    { name: 'group', errorMessage: 'This is a required field.' },
    { name: 'image', errorMessage: 'This is a required field.' },
    { name: 'imageTag', errorMessage: 'This is a required field.' },
    { name: 'command', errorMessage: 'This is a required field.' },
    { name: 'memory', errorMessage: 'This is a required field.' },
  ]
  const errors = {}

  requiredFields.forEach((field) => {
    if (!values[field.name]) {
      errors[field.name] = field.errorMessage
    }
  })

  if (!!values.env) {
    const envErrors = []
    values.env.forEach((e, i) => {
      const envError = {}
      if (!e.name) {
        envError.name = 'Variable name can\'t be blank.'
      } else if (invalidEnv.includes(e.name)) {
        envError.name = 'The name of this variable is illegal.'
      }

      if (!!envError && Object.keys(envError).length > 0) {
        envErrors[i] = envError
      }
    })
    if (envErrors.length) {
      errors.env = envErrors
    }
  }

  return errors
}
