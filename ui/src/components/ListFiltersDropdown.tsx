import * as React from "react"
import { Button, Tooltip, Popover, Position, Card } from "@blueprintjs/core"

const ListFiltersDropdown: React.FunctionComponent<{}> = ({ children }) => (
  <Popover
    minimal
    position={Position.BOTTOM_RIGHT}
    content={<Card className="flotilla-list-filters-card">{children}</Card>}
  >
    <Tooltip content="Show Advanced Filters">
      <Button icon="filter-list" />
    </Tooltip>
  </Popover>
)

export default ListFiltersDropdown
