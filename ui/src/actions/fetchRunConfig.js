import { replace } from 'react-router-redux'
import localforage from 'localforage'
import { localStorageKey } from '../constants/'

export default function fetchRunConfig({ id }) {
  return (dispatch, getState) => {
    localforage.getItem(localStorageKey)
      .then((val) => {
        // After the task config has been fetched from local storage,
        // a few things can happen. If the cluster or envvars are
        // specified in the query, they will take precedence over the
        // local config.
        const { pathname, query } = getState().routing.locationBeforeTransitions
        const newQuery = {
          cluster: !!query.cluster ? query.cluster : (!!val && !!val[id] && !!val[id].cluster) ? val[id].cluster : 'flotilla-adhoc',
        }

        // Yikes.
        if (!!query.env || (!!val && !!val[id] && !!val[id].env)) {
          newQuery.env = !!query.env ? query.env : (!!val && !!val[id] && !!val[id].env) ? val[id].env : null
        }

        dispatch(replace({
          pathname,
          query: newQuery
        }))
      })
  }
}
