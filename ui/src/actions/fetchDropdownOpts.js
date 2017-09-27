import config from '../config'
import { ActionTypes, getApiRoot, imageEndpoint } from '../constants/'

const requestDropdownOpts = () => ({ type: ActionTypes.REQUEST_DROPDOWN_OPTS })
const receiveDropdownOpts = res => ({
  type: ActionTypes.RECEIVE_DROPDOWN_OPTS,
  payload: {
    image: res[0].repositories.map(o => ({ label: o, value: o })),
    group: res[1].groups.map(o => ({ label: o, value: o })),
    cluster: res[2].registered_clusters.map(o => ({ label: o, value: o })),
    tag: res[3].tags.map(o => ({ label: o, value: o }))
  }
})
const receiveDropdownOptsError = error => ({
  type: ActionTypes.RECEIVE_DROPDOWN_OPTS_ERROR,
  payload: { error }
})

export default function fetchDropdownOpts() {
  return (dispatch) => {
    dispatch(requestDropdownOpts())

    const asyncFetchDropdownOpts = async () => {
      try {
        const image = (await fetch(imageEndpoint)).json()
        // Note: need to set a high limit for groups, otherwise will
        // only return 100.
        const group = (await fetch(`${getApiRoot()}/groups?limit=2000`)).json()
        const cluster = (await fetch(config.CLUSTERS_API_ROOT)).json()
        const tag = (await fetch(`${getApiRoot()}/tags?limit=5000`)).json()

        const fetchAll = () => Promise.all([image, group, cluster, tag])
          .then(
            (res) => { dispatch(receiveDropdownOpts(res)) },
            (err) => {
              console.log(err)
            }
          )

        fetchAll()
      } catch (err) {
        dispatch(receiveDropdownOptsError(err))
      }
    }

    asyncFetchDropdownOpts()
  }
}
