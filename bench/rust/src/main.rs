//! Experimental syntax recognizer; intentionally no semantic validation or AST.
use serde::Deserialize;
use std::{collections::{HashMap, HashSet}, env, fs, time::Instant};

#[derive(Deserialize)]
struct Expr {
    op: String,
    #[serde(default)] text: String,
    #[serde(default)] args: Vec<Expr>,
}
#[derive(Deserialize)]
struct Rule { expr: Expr }
#[derive(Clone, Copy)]
struct Token<'a> { kind: u8, text: &'a str }
#[derive(Clone, Copy)]
struct Match { end: usize, ok: bool }
fn yes(end: usize) -> Match { Match { end, ok: true } }
fn no(end: usize) -> Match { Match { end, ok: false } }
fn id_start(b: u8) -> bool { b.is_ascii_alphabetic() || b == b'_' }
fn id_part(b: u8) -> bool { id_start(b) || b.is_ascii_digit() }

fn collect(e: &Expr, literals: &mut HashSet<String>) {
    if e.op == "lit" { literals.insert(e.text.clone()); }
    for a in &e.args { collect(a, literals); }
}
fn lex<'a>(s: &'a str, literals: &[String]) -> Result<Vec<Token<'a>>, String> {
    let mut out = Vec::new();
    let mut i = 0;
    while i < s.len() {
        let r = &s[i..];
        let b = r.as_bytes();
        if b" \t\r\n\x0c".contains(&b[0]) { i += 1; continue; }
        if r.starts_with("//") { i += r.find('\n').unwrap_or(r.len()); continue; }
        if r.starts_with("/*") { i += r[2..].find("*/").ok_or("comment")? + 4; continue; }
        if b[0] == b'\'' || b[0] == b'"' {
            let mut j = 1;
            while j < b.len() && b[j] != b[0] {
                if b[j] == b'\\' {
                    j += 1;
                    if j == b.len() {return Err("escape".into());}
                    if b[j]==b'u' {
                        let first=unicode_unit(&r[j..])?;
                        if (0xd800..=0xdbff).contains(&first) {
                            if j+11>b.len()||&r[j+5..j+7]!="\\u" {return Err("surrogate".into());}
                            let second=unicode_unit(&r[j+6..])?;
                            if !(0xdc00..=0xdfff).contains(&second) {return Err("surrogate".into());}
                            j+=10;
                        } else {if (0xdc00..=0xdfff).contains(&first) {return Err("surrogate".into());}j+=4;}
                    } else if !b"nrtbf\\\'\"".contains(&b[j]) {return Err("escape".into());}
                    j += 1;
                } else { j += r[j..].chars().next().unwrap().len_utf8(); }
            }
            if j == b.len() { return Err("string".into()); }
            j += 1; out.push(Token { kind: 2, text: &r[..j] }); i += j; continue;
        }
        if b[0] == b'^' && b.len() > 1 && id_start(b[1]) {
            let mut j = 2; while j < b.len() && id_part(b[j]) { j += 1; }
            out.push(Token { kind: 1, text: &r[1..j] }); i += j; continue;
        }
        let literal = literals.iter().find(|v| r.starts_with(v.as_str()) && !(id_part(v.as_bytes()[v.len()-1]) && b.len() > v.len() && id_part(b[v.len()])));
        if let Some(v) = literal { out.push(Token { kind: 0, text: &r[..v.len()] }); i += v.len(); continue; }
        if id_start(b[0]) {
            let mut j = 1; while j < b.len() && id_part(b[j]) { j += 1; }
            out.push(Token { kind: 1, text: &r[..j] }); i += j; continue;
        }
        if b[0].is_ascii_digit() {
            let mut j = 1; while j < b.len() && b[j].is_ascii_digit() { j += 1; }
            out.push(Token { kind: 3, text: &r[..j] }); i += j; continue;
        }
        let n = r.chars().next().unwrap().len_utf8(); out.push(Token { kind: 4, text: &r[..n] }); i += n;
    }
    out.push(Token { kind: 5, text: "" }); Ok(out)
}
struct Parser<'a> {
    rules: &'a HashMap<String, Rule>,
    tokens: &'a [Token<'a>],
    memo: HashMap<(&'a str, usize), Match>,
    depth: usize,
    steps: usize,
    limited: bool,
}
impl<'a> Parser<'a> {
    fn sequence(&mut self,args: &'a [Expr],pos: usize)->Match {
        if args.is_empty() {return yes(pos);}
        let first=&args[0];let m=self.eval(first,pos);
        if m.ok {let rest=self.sequence(&args[1..],m.end);if rest.ok {return rest;}}
        if first.op=="repeat"&&first.text=="?" {return self.sequence(&args[1..],pos);}
        no(pos)
    }

    fn rule(&mut self, name: &'a str, pos: usize) -> Match {
        let terminal = match name { "ID" => Some(1), "STRING" => Some(2), "INT" => Some(3), _ => None };
        if let Some(kind) = terminal { return if self.tokens[pos].kind == kind { yes(pos+1) } else { no(pos) }; }
        if name == "SL_COMMENT" || name == "ML_COMMENT" { return no(pos); }
        if let Some(m) = self.memo.get(&(name, pos)) { return *m; }
        self.depth += 1;
        if self.depth > 256 { self.limited = true; self.depth -= 1; return no(pos); }
        let m = if let Some(r) = self.rules.get(name) { self.eval(&r.expr, pos) } else { no(pos) };
        self.depth -= 1; self.memo.insert((name, pos), m); m
    }
    fn eval(&mut self, e: &'a Expr, pos: usize) -> Match {
        self.steps += 1; if self.steps > 20_000_000 { self.limited = true; return no(pos); }
        match e.op.as_str() {
            "lit" => if self.tokens[pos].kind == 0 && self.tokens[pos].text == e.text { yes(pos+1) } else { no(pos) },
            "call" => self.rule(&e.text, pos),
            "ref" => self.rule("ID", pos),
            "action" => yes(pos),
            "assign" => self.eval(&e.args[0], pos),
            "seq" => self.sequence(&e.args, pos),
            "alt" => { let mut best = no(pos); for a in &e.args { let m = self.eval(a, pos); if m.ok && (!best.ok || m.end > best.end) { best = m; } } best },
            "repeat" => {
                let mut at = pos; let mut count = 0;
                loop { let m = self.eval(&e.args[0], at); if feature_boundary(&e.args[0],self.tokens,m) || !m.ok || m.end == at { break; } at = m.end; count += 1; if e.text == "?" { break; } }
                if e.text == "+" && count == 0 { no(pos) } else { yes(at) }
            },
            "unordered" => {
                let mut at = pos; let mut used = 0u64; if e.args.len()>64 {self.limited=true;return no(pos);}
                loop {
                    let mut index = None; let mut best = no(at);
                    for (i,a) in e.args.iter().enumerate() { if used & (1u64<<i)!=0 { continue; } let m = self.eval(a, at); if m.ok && m.end > best.end { index = Some(i); best = m; } }
                    if let Some(i) = index { used |= 1u64<<i; at = best.end; } else { break; }
                }
                for (i,a) in e.args.iter().enumerate() { if used & (1u64<<i)==0 { let m = self.eval(a, at); if !m.ok { return m; } } }
                yes(at)
            },
            _ => panic!("unsupported expression"),
        }
    }
}
fn check(source: &str, rules: &HashMap<String,Rule>, literals: &[String]) -> bool {
    let Ok(tokens) = lex(source, literals) else { return false; };
    let mut p = Parser { rules, tokens: &tokens, memo: HashMap::new(), depth: 0, steps: 0, limited: false };
    let m = p.rule("ContextMappingModel", 0); m.ok && m.end == tokens.len()-1 && !p.limited
}
fn main() {
    let args: Vec<String> = env::args().collect();
    let input = args.get(1).expect("input");
    let n: usize = args.get(2).map(|s| s.parse().unwrap()).unwrap_or(200);
    let source = fs::read_to_string(input).unwrap();
    let rules: HashMap<String,Rule> = serde_json::from_str(include_str!("../../grammar/grammar.json")).unwrap();
    let mut set = HashSet::new(); for r in rules.values() { collect(&r.expr, &mut set); }
    let mut literals: Vec<String> = set.into_iter().collect(); literals.sort_by(|a,b| b.len().cmp(&a.len()).then(a.cmp(b)));
    if !check(&source, &rules, &literals) { std::process::exit(1); }
    let start = Instant::now();
    for _ in 0..n { assert!(std::hint::black_box(check(std::hint::black_box(&source), &rules, &literals))); }
    println!("{{\"language\":\"rust\",\"bytes\":{},\"iterations\":{},\"ns_per_op\":{},\"syntax_only\":true}}",source.len(),n,start.elapsed().as_nanos()/n as u128);
}

fn unicode_unit(s:&str)->Result<u16,String> {
    if s.len()<5||s.as_bytes()[0]!=b'u' {return Err("unicode".into());}
    let digits=s.as_bytes().get(1..5).ok_or("unicode")?;
    if !digits.iter().all(|c|c.is_ascii_hexdigit()) {return Err("unicode".into());}
    u16::from_str_radix(std::str::from_utf8(digits).unwrap(),16).map_err(|_|"unicode".into())
}
fn feature_boundary(e:&Expr,tokens:&[Token],m:Match)->bool {
    if !m.ok||e.op!="seq"||e.args.len()!=2||e.args[0].op!="lit"||e.args[0].text!=","||e.args[1].op!="assign"||e.args[1].text!="entityAttributes" {return false;}
    tokens[m.end].kind==2 || tokens[m.end].kind==0&&["a","an","the"].contains(&tokens[m.end].text)
}
