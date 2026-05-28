-- Partial UNIQUE на (method, ext_id) — защита от двойного зачёта оплаты,
-- когда платёжный вебхук провайдера приходит несколько раз. Применяется
-- только к не-P2P платежам (P2P хранит ext_id = '').
CREATE UNIQUE INDEX IF NOT EXISTS uq_payments_method_extid
    ON payments(method, ext_id)
    WHERE ext_id <> '';
