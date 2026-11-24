# Corvo API contract (vendored)

`openapi.json` is a copy of the public API contract (version 0.3.0) from
`docs/openapi/openapi.json` of corvo-technology/corvo-nfse (also served at
<https://corvo.tech/openapi.json>). `API-CHANGELOG.md` is the contract changelog.

`test/contract_coverage_test.go` loads this file and fails when an operation of
the contract is not implemented by the SDK. To update after an API change:

```bash
curl -fsSL https://corvo.tech/openapi.json -o spec/openapi.json
go test ./test/ -run TestContractCoverage
```
