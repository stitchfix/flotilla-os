import React from "react"
import { mount } from "enzyme"
import Connected, { ClusterSelect } from "../ClusterSelect"
import { ListClustersResponse } from "../../types"
import FlotillaClient from "../../helpers/FlotillaClient"

const mockListClusters = jest.fn()
jest.mock("../../helpers/FlotillaClient", () => {
  return jest.fn().mockImplementation(() => {
    return { listClusters: mockListClusters }
  })
})

describe("ClusterSelect", () => {
  it("renders", () => {
    const wrapper = mount(<Connected value="" onChange={jest.fn()} />)
    expect(true).toBe(true)
  })
})
