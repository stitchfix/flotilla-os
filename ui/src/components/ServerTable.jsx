import React from 'react'
import PropTypes from 'prop-types'
import { FormGroup } from './'

const ServerTable = ({
  queryInputs,
  headers,
  query,
  children,
  onSortChange,
}) => (
  <div className="server-table">
    {
      queryInputs &&
        <div className="flex server-table-inputs">
          {
            queryInputs.map((input, i) => (
              <FormGroup
                key={`query-input-${i}`}
                style={input.style}
                label={input.label}
                input={input.input}
              />
            ))
          }
        </div>
    }
    <div className="table hoverable">
      {
        headers &&
          <div className="thead">
            <div className="tr">
              {
                headers.map((h, i) => {
                  let className = ''
                  if (h.key === query.sort_by) { className += ` active ${query.order}` }
                  if (h.sortable) { className += ' is-sortable' }
                  return (
                    <div
                      key={`th-${i}`}
                      onClick={() => { if (h.sortable) { onSortChange(h.key) } }}
                      className={`th ${className}`}
                      style={h.style}
                    >
                      {h.displayName}
                    </div>
                  )
                })
              }
            </div>
          </div>
      }
      <div className="tbody">
        {children}
      </div>
    </div>
  </div>
)

ServerTable.propTypes = {
  queryInputs: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string.isRequired,
    input: PropTypes.node.isRequired
  })),
  headers: PropTypes.arrayOf(PropTypes.shape({
    displayName: PropTypes.string,
    key: PropTypes.string,
    sortable: PropTypes.bool,
  })),
  query: PropTypes.object,
  onSortChange: PropTypes.func,
}

export default ServerTable
