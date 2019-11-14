import * as React from "react"
import { get, isArray } from "lodash"
import Creatable from "react-select/lib/Creatable"
import Request from "./Request"
import { ListClustersResponse, SelectOption, SelectProps } from "../types"
import api from "../api"
import * as helpers from "../helpers/selectHelpers"

/**
 * ClusterSelect allows users to select an ECS cluster on which to run a
 * particular task. This component hits the `/clusters` endpoint and renders
 * the results into a React Select component.
 */
export const ClusterSelect: React.FunctionComponent<SelectProps & {
  options: SelectOption[]
}> = props => {
  return (
    <Creatable<SelectOption>
      value={helpers.stringToSelectOpt(props.value)}
      options={props.options}
      isClearable
      onChange={option => {
        props.onChange(helpers.preprocessSelectOption(option))
      }}
      styles={helpers.selectStyles}
      theme={helpers.selectTheme}
      isDisabled={props.isDisabled}
    />
  )
}

const Connected: React.FunctionComponent<SelectProps> = props => (
  <Request<ListClustersResponse, {}> requestFn={api.listClusters}>
    {res => {
      let options = get(res, ["data", "clusters"], [])

      // If there's an error fetching available clusters, set the options to
      // an empty array.
      if (!isArray(options)) options = []
      return (
        <ClusterSelect
          options={options.map(helpers.stringToSelectOpt)}
          value={props.value}
          onChange={props.onChange}
          isDisabled={props.isDisabled}
        />
      )
    }}
  </Request>
)

export default Connected
