import React from "react"
import { QueryParamsFieldText } from "../Field/FieldText"
import { QueryParamsFieldSelect } from "../Field/FieldSelect"
import QueryParamsKVField from "../Field/QueryParamsKVField"
import { asyncDataTableFilters, IAsyncDataTableFilterProps } from "../../.."

class AsyncDataTableFilter extends React.PureComponent<
  IAsyncDataTableFilterProps
> {
  render() {
    const { description, displayName, name, type, filterProps } = this.props

    const sharedProps = {
      label: displayName,
      name,
      description,
      ...filterProps,
    }

    switch (type) {
      case asyncDataTableFilters.KV:
        return (
          <QueryParamsKVField {...sharedProps} isKeyRequired isValueRequired />
        )
      case asyncDataTableFilters.SELECT:
        return <QueryParamsFieldSelect {...sharedProps} {...this.props} />
      case asyncDataTableFilters.INPUT:
      default:
        return <QueryParamsFieldText {...sharedProps} shouldDebounce />
    }
  }
}

export default AsyncDataTableFilter
