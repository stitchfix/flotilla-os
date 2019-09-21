import * as React from "react"
import { get, isArray } from "lodash"
import Select from "react-select"
import Request from "./Request"
import { ListClustersResponse, SelectOption, SelectProps } from "../types"
import api from "../api"
import * as helpers from "../helpers/selectHelpers"

/**
 * ClusterSelect allows users to select an ECS cluster on which to run a
 * particular task. This component hits the `/clusters` endpoint and renders
 * the results into a React Select component.
 */
export const ClusterSelect: React.FunctionComponent<
  SelectProps & { options: SelectOption[] }
> = props => {
  return (
    <Select<SelectOption>
      value={helpers.stringToSelectOpt(props.value)}
      options={props.options}
      onChange={option => {
        props.onChange(helpers.preprocessSelectOption(option))
      }}
    />
  )
}

const Connected: React.FunctionComponent<SelectProps> = props => (
  <Request<ListClustersResponse, {}> requestFn={api.listClusters}>
    {res => {
      let options = get(res, ["data", "clusters"], [])
      if (!isArray(options)) options = []
      return (
        <ClusterSelect
          options={options.map(helpers.stringToSelectOpt)}
          value={props.value}
          onChange={props.onChange}
        />
      )
    }}
  </Request>
)

export default Connected
