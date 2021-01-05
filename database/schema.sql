DROP TABLE IF EXISTS public.sales;
CREATE TABLE IF NOT EXISTS public.sales {
    sale_id int PRIMARY KEY,
    offer_id int,
    seller_id int,
    price int,
    name varchar(200),
    quantity int
};

CREATE INDEX sale_pair_index ON public.sales(offer_id, seller_id);