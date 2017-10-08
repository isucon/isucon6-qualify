#[macro_use]
extern crate iron;
extern crate iron_sessionstorage;
#[macro_use]
extern crate router;

use iron::prelude::*;
use iron::status;
use iron::modifiers::Redirect;

use router::Router;

use iron_sessionstorage::traits::*;
use iron_sessionstorage::SessionStorage;
use iron_sessionstorage::backends::SignedCookieBackend;

struct UserName {
    name: String
}

impl iron_sessionstorage::Value for UserName {
    fn get_key() -> &'static str { "username" }
    fn into_raw(self) -> String { self.name }
    fn from_raw(value: String) -> Option<Self> {
        if value.is_empty() {
            None
        } else {
            Some(UserName { name: value })
        }
    }
}

/*
fn set_name(req: &Request) {
    let mut value = match try!(req.session().get::<UserId>()) {
        Some(user_id)
    }
}
*/

/*fn login_handler(req: &mut Request) -> IronResult<Response> {

}*/

fn top_handler(req: &mut Request) -> IronResult<Response> {
    let userName = req.session().get::<UserName>().ok().and_then(|x| x);
    let name;
    match userName {
        Some(un) => { name = un.name }
        None => { name = String::from("nobody") }
    }
    try!(req.session().set(UserName { name: String::from("somebody") }));
    Ok(Response::with((status::Found, format!("Hello, {}, reload to become \"somebody\"", name))))
}

fn main() {
    let mysecret = b"foobar".to_vec();
    let mut router = Router::new();
    router.get("/", top_handler, "top_handler");
    let mut ch = Chain::new(router);
    ch.link_around(SessionStorage::new(SignedCookieBackend::new(mysecret)));
    let _res = Iron::new(ch).http("localhost:8080");
}
