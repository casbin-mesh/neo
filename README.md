
<p align="center"><img src="./casbin-neo.png" width="370"></p>
<p align="center">
<b>A Casbin-compatible engine</b>
</p>

# Casbin NEO
Casbin NEO(neo for new engine option), A Casbin-compatible engine.
In this project, we would go to restructure the Casbin storage layer, which will employ the column-oriented store, and supports transactions executing under snapshot isolation level.

Furthermore, we were planning to explore ideas from state-of-art systems and research, such as the query on compressed data, query compilation, fast serializable snapshot isolation, etc.

**NOTE: This project is still under development.**

<p>
  <a href="https://goreportcard.com/report/github.com/casbin-mesh/neo">
    <img src="https://goreportcard.com/badge/github.com/casbin-mesh/neo">
  </a>
  <a href="https://godoc.org/github.com/casbin-mesh/neo">
    <img src="https://godoc.org/github.com/casbin-mesh/neo?status.svg" alt="GoDoc">
  </a>
    <img src="https://github.com/casbin-mesh/neo/workflows/Go/badge.svg?branch=main"/>
</p>



## Documentation
All documents were located in [docs](/docs) directory.

## License
This project is licensed under the [Apache 2.0 license](/LICENSE).
