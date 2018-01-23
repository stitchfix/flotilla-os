import React from 'react'

export default function EnvContainer({ env }) {
  return (
    <div
      className="flex ff-cn a-fs code"
      style={{
        width: '100%',
      }}
    >
      {
        !!env && env.map((v, i) => (
          <div
            className="flex ff-cn j-fs"
            style={{ marginTop: 16, width: '100%' }}
            key={`env-${i}`}
          >
            <div style={{ flex: 0.3, marginRight: 8, color: '#777777', fontSize: 12 }}>
              {v.name}
            </div>
            <div style={{ fontSize: 14, color: '#c2c2c2' }}>{v.value}</div>
          </div>
        ))
      }
    </div>
  )
}
