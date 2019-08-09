import * as React from "react"
import { Button, ButtonGroup } from "@blueprintjs/core"

type Props = {
  totalPages: number
  updatePage: (n: number) => void
  currentPage: number
}

const Pagination: React.FunctionComponent<Props> = ({
  totalPages,
  updatePage,
  currentPage,
}) => {
  return (
    <ButtonGroup>
      <Button
        onClick={() => {
          updatePage(currentPage - 1)
        }}
        disabled={currentPage === 1}
      >
        Previous Page
      </Button>
      <Button
        onClick={() => {
          updatePage(currentPage + 1)
        }}
        disabled={currentPage === totalPages}
      >
        Next Page
      </Button>
    </ButtonGroup>
  )
}

export default Pagination
