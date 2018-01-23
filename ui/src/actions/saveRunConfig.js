import localforage from 'localforage'
import { localStorageKey } from '../constants/'

export default function saveRunConfig({ cluster, env, taskID }) {
  return () => {
    localforage.getItem(localStorageKey)
      .then((runConfig) => {
        let _runConfig

        if (runConfig === null) {
          _runConfig = {
            [taskID]: {
              cluster,
              // Note: environment variables are coerced to the same
              // signature as in url query for easier parsing later on
              env: env ? env.map(e => `${e.name}|${e.value}`) : [],
              id: taskID
            }
          }
        } else {
          _runConfig = {
            ...runConfig,
            [taskID]: {
              cluster,
              id: taskID
            }
          }

          if (!!env) { _runConfig.env = env.map(e => `${e.name}|${e.value}`) }
        }

        localforage.setItem(localStorageKey, _runConfig)
          .then((val) => {
            console.info(`[Flotilla UI] Run config for task [${taskID}] saved to local storage.`)
            console.log(val)
          })
      })
  }
}
