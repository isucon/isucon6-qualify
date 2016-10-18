'use strict';
const router = require('koa-router')();
const mysql = require('promise-mysql');
const crypto = require('crypto');
const axios = require('axios');
const ejs = require('ejs');

let _config;
const config = (key) => {
  if (!_config) {
    _config = {
      dsn: process.env.ISUDA_DSN || 'dbi:mysql:db=isuda',
      dbHost: process.env.ISUDA_DB_HOST || 'localhost',
      dbPort: process.env.ISUDA_DB_PORT || 3306,
      dbName: process.env.ISUDA_DB_NAME || 'isuda',
      dbUser: process.env.ISUDA_DB_USER || 'root',
      dbPassword: process.env.ISUDA_DB_PASSWORD || '',
      isutarOrigin: process.env.ISUTAR_ORIGIN || 'http://localhost:5001',
      isupamOrigin: process.env.ISUPAM_ORIGIN || 'http://localhost:5050',
    };
  }
  if (!_config.hasOwnProperty(key)) {
    throw `config value of ${key} undefined`;
  }
  return _config[key];
};

// SEE ALSO: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/encodeURIComponent
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
    host: config('dbHost'),
    port: config('dbPort'),
    user: config('dbUser'),
    password: config('dbPassword'),
    database: config('dbName'),
    connectionLimit: 1,
    charset: 'utf8mb4'
  });
  await ctx.dbh.query("SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'");
  await ctx.dbh.query("SET NAMES utf8mb4");

  return ctx.dbh;
};

const setName = async (ctx) => {
  ctx.state = {};
  const db = await dbh(ctx);
  const userId = ctx.session.userId;
  if (userId != null) {
    const users = await db.query('SELECT name FROM user WHERE id = ?', [userId.toString()]);
    if (users.length > 0) {
      ctx.state.user_name = users[0].name;
    } else {
      ctx.status = 403;
      return false;
    }
  }
  return true;
};

const authenticate = (ctx) => {
  if (ctx.session.userId == null) {
    ctx.status = 403;
    return false;
  }
  return true;
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
  await db.query('DELETE FROM entry WHERE id > 7101');
  const origin = config('isutarOrigin');
  const res = await axios.get(`${origin}/initialize`);
  ctx.body = {
    result: 'ok',
  };
});

router.get('', async (ctx, next) => {
  if (!await setName(ctx)) {
    return;
  }
  const perPage = 10;
  const page = parseInt(ctx.query.page) || 1;

  const db = await dbh(ctx);
  const entries = await db.query('SELECT * FROM entry ORDER BY updated_at DESC LIMIT ? OFFSET ?', [perPage, perPage * (page - 1)])
  for (let entry of entries) {
    entry.html = await htmlify(ctx, entry.description);
    entry.stars = await loadStars(ctx, entry.keyword);
  }

  const totalEntries = await db.query('SELECT COUNT(*) AS `count` FROM entry');
  const lastPage = Math.ceil(totalEntries[0].count / perPage);
  const pages = [];
  for (let i = Math.max(1, page - 5); i <= Math.min(lastPage, page + 5); i++) {
    pages.push(i);
  }

  ctx.state.entries = entries;
  ctx.state.page = page;
  ctx.state.lastPage = lastPage;
  ctx.state.pages = pages;

  await ctx.render('index', {
  });
});

router.get('robots.txt', async (ctx, next) => {
  ctx.status = 404;
});

router.post('keyword', async (ctx, next) => {
  if (!await setName(ctx)) {
    return;
  }
  if (!authenticate(ctx)) {
    return;
  }
  const keyword = ctx.request.body.keyword || '';
  if (keyword.length === 0) {
    ctx.status = 400;
    ctx.body = "'keyword' required";
  }
  const userId = ctx.session.userId;
  const description = ctx.request.body.description;

  if (await isSpamContents(description) || await isSpamContents(keyword)) {
    ctx.status = 400;
    ctx.body = 'SPAM!';
    return;
  }

  const db = await dbh(ctx);
  await db.query(
    'INSERT INTO entry (author_id, keyword, description, created_at, updated_at) ' +
    'VALUES (?, ?, ?, NOW(), NOW()) ' +
    'ON DUPLICATE KEY UPDATE ' +
    'author_id = ?, keyword = ?, description = ?, updated_at = NOW()',
    [
      userId, keyword, description, userId, keyword, description
    ]);

  await ctx.redirect('/');

});

router.get('register', async (ctx, next) => {
  if (!await setName(ctx)) {
    return;
  }
  ctx.state.action = 'register';
  await ctx.render('authenticate', {
  });
});

router.post('register', async (ctx, next) => {
  const name = ctx.request.body.name;
  const pw   = ctx.request.body.password;
  if (name === '' || pw === '') {
    ctx.status = 400;
    return;
  }
  const userId = await register(await dbh(ctx), name, pw);
  ctx.session.userId = userId;
  await ctx.redirect('/');
});

const register = async (db, user, pass) => {
  const salt = await randomString(10);
  const sha1 = crypto.createHash('sha1');
  sha1.update(salt + pass);
  await db.query('INSERT INTO user (name, salt, password, created_at) VALUES (?, ?, ?, NOW())', [user, salt, sha1.digest('hex')]);
  const row = await db.query("SELECT LAST_INSERT_ID() as lastInsertId ");
  return row[0].lastInsertId;
};

const randomString = (size) => {
  return new Promise((resolve, reject) => {
    crypto.randomBytes(size, (err, buf) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(buf.toString('hex'));
    });
  });
}

router.get('login', async (ctx, next) => {
  if (!await setName(ctx)) {
    return;
  }
  ctx.state.action = 'login';
  await ctx.render('authenticate', {});
});

router.post('login', async (ctx, next) => {
  const name = ctx.request.body.name;
  const db = await dbh(ctx);
  const rows = await db.query('SELECT * FROM user WHERE name = ?', [name]);
  if (rows.length === 0) {
    ctx.status = 403;
    return;
  }
  const sha1 = crypto.createHash('sha1');
  sha1.update(rows[0].salt + ctx.request.body.password);
  const sha1Digest = sha1.digest('hex');
  if (rows[0].password != sha1Digest) {
    ctx.status = 403;
    return;
  }
  ctx.session.userId = rows[0].id;
  await ctx.redirect('/');
});

router.get('logout', async (ctx, next) => {
  ctx.session.userId = null;
  await ctx.redirect('/');
});

router.get('keyword/:keyword', async (ctx, next) => {
  if (!await setName(ctx)) {
    return;
  }
  const keyword = ctx.params.keyword;
  if (!keyword) {
    ctx.status = 400;
    return;
  }
  const db = await dbh(ctx);
  const entries = await db.query('SELECT * FROM entry WHERE keyword = ?', [keyword]);
  if (entries.length === 0) {
    ctx.status = 404;
    return;
  }
  ctx.state.entry = entries[0];
  ctx.state.entry.html = await htmlify(ctx, entries[0].description);
  ctx.state.entry.stars = await loadStars(ctx, keyword);
  await ctx.render('keyword');
});

router.post('keyword/:keyword', async (ctx, next) => {
  if (!await setName(ctx)) {
    return;
  }
  if (!authenticate(ctx)) {
    return;
  }
  const keyword = ctx.params.keyword;
  if ( !keyword ) {
    ctx.status = 400;
    return;
  }
  const del = ctx.request.body.delete;
  if ( !ctx.request.body.delete ) {
    ctx.status = 400;
    return;
  }

  const db = await dbh(ctx);
  const entries = await db.query('SELECT * FROM entry WHERE keyword = ?', [keyword]);
  if (entries.length == 0) {
    ctx.status = 404;
    return;
  }

  await db.query('DELETE FROM entry WHERE keyword = ?', [keyword]);

  await ctx.redirect('/');
});

const htmlify = async (ctx, content) => {
  if (content == null) {
    return '';
  }

  const db = await dbh(ctx);
  const keywords = await db.query('SELECT * FROM entry ORDER BY CHARACTER_LENGTH(keyword) DESC');
  const key2sha = new Map();
  const re = new RegExp(keywords.map((keyword) => escapeRegExp(keyword.keyword)).join('|'), 'g');
  let result = content.replace(re, (keyword) => {
    const sha1 = crypto.createHash('sha1');
    sha1.update(keyword);
    let sha1hex = `isuda_${sha1.digest('hex')}`;
    key2sha.set(keyword, sha1hex);
    return sha1hex;
  });
  for (let kw of key2sha.keys()) {
    const url = `/keyword/${RFC3986URIComponent(kw)}`;
    const link = `<a href=${url}>${ejs.escapeXML(kw)}</a>`;
    result = result.replace(new RegExp(escapeRegExp(key2sha.get(kw)), 'g'), link);
  }
  result = result.replace(/\n/g, "<br />\n");

  return result;
};

const escapeRegExp  = (string) => {
    return string.replace(/([.*+?^=!:${}()|[\]\/\\])/g, "\\$1");
}

const escapeHtml = (string) => {
};

const loadStars = async (ctx, keyword) => {
  const origin = config('isutarOrigin');
  const url = `${origin}/stars`;
  const res = await axios.get(url, {params: {keyword: keyword}});
  return res.data.stars;
};

const isSpamContents = async (content) => {
  const res = await axios.post(config('isupamOrigin'), `content=${encodeURIComponent(content)}`);
  return !res.data.valid;
};

module.exports = router;
