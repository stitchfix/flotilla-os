import { Tooltip, Colors } from "@blueprintjs/core"


const isLessThanPct = (x: number, y: number, pct: number): boolean => {
    if (x < pct * y) return true
    return false
  }
  
const ResourceUsageValue: React.FC<{
    requested: number | undefined | null
    actual: number | undefined | null
    requestedName: string
    actualName: string
  }> = ({ requested, actual, requestedName, actualName }) => {
    if (!requested) {
      return <span>-</span>
    }
  
    if (!actual) {
      return <span>{requested}</span>
    }
  
    return (
      <div>
        <Tooltip content={actualName}>
          <span
            style={{
              color:
                actual && isLessThanPct(actual, requested, 0.5)
                  ? Colors.RED5
                  : "",
            }}
          >
            {actual}
          </span>
        </Tooltip>{" "}
        / <Tooltip content={requestedName}>{requested}</Tooltip>
      </div>
    )
  }

  export default ResourceUsageValue