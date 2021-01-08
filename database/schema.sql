DROP TABLE IF EXISTS sales;
CREATE TABLE IF NOT EXISTS sales (
    sale_id int PRIMARY KEY,
    offer_id int,
    seller_id int,
    price int,
    name varchar(200),
    quantity int
);

CREATE INDEX sale_pair_index ON sales(offer_id, seller_id);

INSERT INTO sales (offer_id, seller_id, price, name, quantity) VALUES (1, 1, 100, "Test sale", 1);