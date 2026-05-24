/**
 * Example: Express app with dusk deprecation middleware.
 * Run: node examples/express_app.js
 */
const express = require('express')
const { createDusk } = require('dusk-js')

const app = express()
const dusk = createDusk('./dusk.yaml')

app.use(dusk)

app.get('/api/v1/users', (req, res) => {
  res.json({ users: [], note: 'deprecated — use /api/v2/users' })
})

app.get('/api/v2/users', (req, res) => {
  res.json({ users: [], page: 1, total: 0 })
})

app.get('/api/v1/orders/:id', (req, res) => {
  res.json({ order_id: req.params.id })
})

app.listen(3000, () => {
  console.log('Express app running on http://localhost:3000')
})
