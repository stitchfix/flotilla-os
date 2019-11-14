import * as React from "react"
import { get, isArray } from "lodash"
import Creatable from "react-select/lib/Creatable"
import Request from "./Request"
import { ListTagsResponse, SelectOption, MultiSelectProps } from "../types"
import api from "../api"
import * as helpers from "../helpers/selectHelpers"

export const TagsSelect: React.FunctionComponent<MultiSelectProps & {
  options: SelectOption[]
}> = props => (
  <Creatable<SelectOption>
    isMulti
    value={props.value.map(helpers.stringToSelectOpt)}
    options={props.options}
    onChange={options => {
      props.onChange(helpers.preprocessMultiSelectOption(options))
    }}
    styles={helpers.selectStyles}
    theme={helpers.selectTheme}
    closeMenuOnSelect={false}
    isDisabled={props.isDisabled}
  />
)

const ConnectedTagsSelect: React.FunctionComponent<MultiSelectProps> = props => (
  <Request<ListTagsResponse, {}> requestFn={api.listTags}>
    {res => {
      let options = get(res, ["data", "tags"], [])
      if (!isArray(options)) options = []
      return (
        <TagsSelect
          value={props.value || []}
          options={options.map(helpers.stringToSelectOpt)}
          onChange={props.onChange}
          isDisabled={props.isDisabled}
        />
      )
    }}
  </Request>
)

export default ConnectedTagsSelect
