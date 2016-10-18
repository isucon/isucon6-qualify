const Koa = require('koa');
const session = require('koa-session');
const app = new Koa();
const router = require('koa-router')();
const views = require('koa-views');
const co = require('co');
const convert = require('koa-convert');
const json = require('koa-json');
const onerror = require('koa-onerror');
const bodyparser = require('koa-bodyparser')({ formLimit: '5mb' });
const logger = require('koa-logger');

const isuda = require('./routes/isuda');

// middlewares
app.use(convert(bodyparser));
app.use(convert(json()));
app.use(convert(logger()));
app.keys = ['tonymoris'];
app.use(convert(session(app, { key: 'isuda_session', maxAge: 3600 * 24 })));
app.use(require('koa-static')(__dirname + '/public'));

app.use(views(__dirname + '/views', {
  extension: 'ejs'
}));

// logger
app.use(async (ctx, next) => {
  const start = new Date();
  await next();
  const ms = new Date() - start;
  console.log(`[isuda] ${ctx.method} ${ctx.url} - ${ms}ms`);
});

router.use('/', isuda.routes(), isuda.allowedMethods());

app.use(router.routes(), router.allowedMethods());
// response

app.on('error', function(err, ctx){
  console.log(err)
  logger.error('server error', err, ctx);
});


module.exports = app;
