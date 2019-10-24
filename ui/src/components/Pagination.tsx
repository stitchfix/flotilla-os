import * as React from "react"
import { Button, ButtonGroup } from "@blueprintjs/core"

export type Props = {
  updatePage: (n: number) => void
  currentPage: number
  numItems: number
  pageSize: number
  isLoading: boolean
}

const Pagination: React.FunctionComponent<Props> = ({
  numItems,
  pageSize,
  updatePage,
  currentPage,
  isLoading,
}) => {
  const isFirstPage = currentPage === 1
  const isLastPage = currentPage * pageSize >= numItems
  return (
    <ButtonGroup>
      <Button
        onClick={() => {
          updatePage(currentPage - 1)
        }}
        disabled={isFirstPage || isLoading}
        loading={isLoading}
        icon="chevron-left"
      />
      <Button
        onClick={() => {
          updatePage(currentPage + 1)
        }}
        disabled={isLastPage || isLoading}
        loading={isLoading}
        icon="chevron-right"
      />
    </ButtonGroup>
  )
}

export default Pagination
