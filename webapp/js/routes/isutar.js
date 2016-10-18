'use strict';
const router = require('koa-router')();
const mysql = require('promise-mysql');
const axios = require('axios');

const RFC3986URIComponent = (str) => {
    return encodeURIComponent(str).replace(/[!'()*]/g, (c) => {
          return '%' + c.charCodeAt(0).toString(16);
    });
};

const dbh = async (ctx) => {
  if (ctx.dbh) {
    return ctx.dbh;
  }

  ctx.dbh = mysql.createPool({
    host: process.env.ISUTAR_DB_HOST || 'localhost',
    port: process.env.ISUTAR_DB_PORT || 3306,
    user: process.env.ISUTAR_DB_USER || 'root',
    password: process.env.ISUTAR_DB_PASSWORD || '',
    database: 'isutar',
    connectionLimit: 1,
    charset: 'utf8mb4'
  });
  await ctx.dbh.query("SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'");
  await ctx.dbh.query("SET NAMES utf8mb4");

  return ctx.dbh;
};

router.use(async (ctx, next) => {
  await next();
  if (ctx.dbh) {
    await ctx.dbh.end();
    ctx.dbh = null;
  }
});

router.get('initialize', async (ctx, next) => {
  const db = await dbh(ctx);
  await db.query('TRUNCATE star');
  ctx.body = {
    result: 'ok',
  };
});

router.get('stars', async (ctx, next) => {
  const db = await dbh(ctx);
  const stars =  await db.query('SELECT * FROM star WHERE keyword = ?', [ctx.query.keyword]);
  ctx.body = {
    stars: stars,
  };
});

router.post('stars', async (ctx, next) => {
  const db = await dbh(ctx);
  const keyword = ctx.query.keyword || ctx.request.body.keyword;

  const origin = process.env.ISUDA_ORIGIN || 'http://localhost:5000';
  const url = `${origin}/keyword/${RFC3986URIComponent(keyword)}`;

  try {
    const res = await axios.get(url);
  } catch (err) {
    console.log(err);
    ctx.status = 404;
    return;
  }

  await db.query('INSERT INTO star (keyword, user_name, created_at) VALUES (?, ?, NOW())', [
    keyword, ctx.query.user || ctx.request.body.user
  ]);

  ctx.body = {
    result: 'ok',
  };
});

module.exports = router;
