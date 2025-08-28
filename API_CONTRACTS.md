# 📡 Contratos de API ORDM

## Endpoints de Sincronização

### POST /api/sync/block
```json
{
  "blocks": [
    {
      "hash": "block_hash",
      "parent_hash": "parent_hash",
      "number": 1234,
      "miner_id": "miner_address",
      "transactions": [],
      "pow_proof": "proof_data",
      "signature": "miner_signature"
    }
  ],
  "miner_id": "miner_address",
  "batch_id": "unique_batch_id",
  "timestamp": 1640995200
}
```

### GET /api/sync/status
```json
{
  "status": "syncing",
  "last_sync": "2024-01-01T00:00:00Z",
  "pending_blocks": 5,
  "synced_blocks": 1234
}
```

## Endpoints de Validação

### POST /api/validator/vote
```json
{
  "block_hash": "block_hash",
  "validator_id": "validator_address",
  "vote": true,
  "stake_amount": 1000,
  "signature": "validator_signature"
}
```

### GET /api/validator/stats
```json
{
  "validator_id": "validator_address",
  "stake_amount": 1000,
  "rewards_earned": 50,
  "blocks_validated": 100,
  "reputation": 95.5
}
```
