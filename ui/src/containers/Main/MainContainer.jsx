import React from 'react'
import { Link } from 'react-router'
import { MainNav } from '../'

const tabs = [
  {
    displayName: 'Task Definitions',
    path: '/tasks'
  },
  {
    displayName: 'Active Runs',
    path: '/runs'
  },
]

export default function MainContainer({ location, children }) {
  return (
    <div className="view-container">
      <MainNav location={location} />
      <div className="view">
        <div className="layout-standard">
          <div className="tab-container">
            <div className="tab-container-list">
              {
                tabs.map((tab, i) => {
                  const isActive = location.pathname.endsWith(tab.path)
                  return (
                    <Link
                      className={`tab-container-list-element ${isActive ? 'active' : ''}`}
                      to={{ pathname: tab.path }}
                      key={`tab-link-${i}`}
                    >
                      {tab.displayName}
                    </Link>
                  )
                })
              }
            </div>
            <div className="tab-container-content">
              {children}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
