-- Скрипт для PostgreSQL.
-- Добавляет отель, который ожидается в Postman-тестах:
-- hotelUid = 049161bb-badd-4fa8-9d90-87c9a82b0668
-- Если имена таблицы/колонок в вашей БД отличаются — поправьте их.

BEGIN;

CREATE TABLE IF NOT EXISTS hotels
(
    id        SERIAL PRIMARY KEY,
    hotel_uid uuid         NOT NULL UNIQUE,
    name      VARCHAR(255) NOT NULL,
    country   VARCHAR(80)  NOT NULL,
    city      VARCHAR(80)  NOT NULL,
    address   VARCHAR(255) NOT NULL,
    stars     INT,
    price     INT          NOT NULL
);


INSERT INTO hotels (
    hotel_uid,
    name,
    country,
    city,
    address,
    stars,
    price
)
VALUES (
    '049161bb-badd-4fa8-9d90-87c9a82b0668',
    'Ararat Park Hyatt Moscow',
    'Россия',
    'Москва',
    'Неглинная ул., 4',
    5,
    10000
)
ON CONFLICT (hotel_uid) DO UPDATE
SET
    name    = EXCLUDED.name,
    country = EXCLUDED.country,
    city    = EXCLUDED.city,
    address = EXCLUDED.address,
    stars   = EXCLUDED.stars,
    price   = EXCLUDED.price;

COMMIT;
