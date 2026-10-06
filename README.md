# Calculadora científica (Go + Gin)

API RESTful de calculadora científica. Cada cálculo é um recurso (`/calculations`)
que pode ser criado, consultado, listado e removido.

## Executar

```bash
go run .            # porta 8080 (ou defina PORT)
go test ./...
```

## Endpoints

| Método | Rota                          | Descrição                              | Sucesso |
|--------|-------------------------------|----------------------------------------|---------|
| GET    | `/health`                     | Verificação de saúde                   | 200     |
| GET    | `/api/v1/operations`          | Lista as operações suportadas          | 200     |
| GET    | `/api/v1/operations/:name`    | Detalha uma operação                   | 200     |
| POST   | `/api/v1/calculations`        | Executa e salva um cálculo             | 201     |
| GET    | `/api/v1/calculations`        | Histórico de cálculos                  | 200     |
| GET    | `/api/v1/calculations/:id`    | Consulta um cálculo                    | 200     |
| DELETE | `/api/v1/calculations/:id`    | Remove um cálculo                      | 204     |

### Exemplo

```bash
curl -i -X POST localhost:8080/api/v1/calculations \
  -H 'Content-Type: application/json' \
  -d '{"operation":"sin","operands":[30],"angle_unit":"deg"}'
```

```http
HTTP/1.1 201 Created
Location: /api/v1/calculations/1

{"id":1,"operation":"sin","operands":[30],"angle_unit":"deg","result":0.49999999999999994,"created_at":"..."}
```

`angle_unit` (`rad` padrão ou `deg`) vale para `sin`, `cos`, `tan`, `asin`, `acos` e `atan`.

## Operações

- **Binárias** (`operands: [x, y]`): `add`, `subtract`, `multiply`, `divide`, `mod`,
  `power`, `root` (raiz y-ésima de x), `log` (log de x na base y)
- **Unárias** (`operands: [x]`): `sqrt`, `cbrt`, `abs`, `reciprocal`, `factorial`, `exp`,
  `ln`, `log10`, `log2`, `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `sinh`, `cosh`, `tanh`

## Erros

Formato: `{"error": {"code": "...", "message": "..."}}`

| Status | Códigos                                                                 |
|--------|-------------------------------------------------------------------------|
| 400    | `invalid_body`, `invalid_id`                                            |
| 404    | `operation_not_found`, `calculation_not_found`                          |
| 422    | `unknown_operation`, `wrong_operand_count`, `division_by_zero`, `domain_error`, `overflow`, `invalid_angle_unit` |

O histórico fica em memória e é perdido ao reiniciar o servidor.
