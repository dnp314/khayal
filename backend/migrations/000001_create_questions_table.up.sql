CREATE TABLE IF NOT EXISTS questions (
id bigserial PRIMARY KEY,
created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
title text NOT NULL,
description text 
);