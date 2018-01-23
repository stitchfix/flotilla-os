import React from 'react'
import PropTypes from 'prop-types'
import { withRouter } from 'react-router'
import { FormGroup } from './'

const ErrorView = ({ error, router }) => (
  <div className="error-view">
    <div className="section-container task-definition">
      <div className="section-header">
        <div className="section-header-text error-message">Error</div>
      </div>
      <div>
        <FormGroup isStatic label="Error Message">
          {error.response.statusText}
        </FormGroup>
        <FormGroup isStatic label="Error Status Code">
          {error.response.status || '-'}
        </FormGroup>
      </div>
    </div>
    <button
      className="button button-primary"
      onClick={() => { router.push('/') }}
    >
      Go to main page
    </button>
  </div>
)

ErrorView.propTypes = {
  error: PropTypes.shape({
    response: PropTypes.shape({
      status: PropTypes.number,
      statusText: PropTypes.string,
    })
  })
}

export default withRouter(ErrorView)
