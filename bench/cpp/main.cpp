// Experimental C++20 CML syntax recognizer; no AST or semantic validation.
#include <algorithm>
#include <chrono>
#include <fstream>
#include <iostream>
#include <iterator>
#include <string>
#include <string_view>
#include <unordered_map>
#include <vector>
#include <cstdint>
struct Expr {std::string op,text;std::vector<int> args;};
struct Token {int kind;std::string_view text;};
struct Match {size_t end;bool ok;};
#include "grammar.inc"
bool start(unsigned char c){return c=='_'||(c>='a'&&c<='z')||(c>='A'&&c<='Z');}
bool part(unsigned char c){return start(c)||(c>='0'&&c<='9');}
bool lex(std::string_view s,std::vector<Token>& out){
 size_t i=0;
 while(i<s.size()){
  auto r=s.substr(i);unsigned char c=r[0];
  if(std::string_view(" \t\r\n\f").find(c)!=std::string_view::npos){++i;continue;}
  if(r.starts_with("//")){auto n=r.find('\n');i+=n==r.npos?r.size():n;continue;}
  if(r.starts_with("/*")){auto n=r.substr(2).find("*/");if(n==r.npos)return false;i+=n+4;continue;}
  if(c=='\''||c=='"'){
   size_t j=1;
   while(j<r.size()&&r[j]!=c){
    if(r[j]=='\\'){
     if(++j==r.size())return false;
     if(r[j]=='u'){
      auto unit=[](std::string_view v)->int {if(v.size()<5||v[0]!='u')return -1;int n=0;for(size_t k=1;k<5;++k){unsigned char x=v[k];int h=x>='0'&&x<='9'?x-'0':x>='a'&&x<='f'?x-'a'+10:x>='A'&&x<='F'?x-'A'+10:-1;if(h<0)return -1;n=n*16+h;}return n;};
      int first=unit(r.substr(j));if(first<0)return false;
      if(first>=0xd800&&first<=0xdbff){if(j+11>r.size()||r.substr(j+5,2)!="\\u")return false;int second=unit(r.substr(j+6));if(second<0xdc00||second>0xdfff)return false;j+=10;}
      else {if(first>=0xdc00&&first<=0xdfff)return false;j+=4;}
     }else if(std::string_view("nrtbf\\\'\"").find(r[j])==r.npos)return false;
    }
    ++j;
   }
   if(j==r.size())return false;++j;out.push_back({2,r.substr(0,j)});i+=j;continue;
  }
  if(c=='^'&&r.size()>1&&start(r[1])){size_t j=2;while(j<r.size()&&part(r[j]))++j;out.push_back({1,r.substr(1,j-1)});i+=j;continue;}
  bool matched=false;
  for(auto& lit:literals)if(r.starts_with(lit)&&!(part(lit.back())&&r.size()>lit.size()&&part(r[lit.size()]))){out.push_back({0,r.substr(0,lit.size())});i+=lit.size();matched=true;break;}
  if(matched)continue;
  if(start(c)){size_t j=1;while(j<r.size()&&part(r[j]))++j;out.push_back({1,r.substr(0,j)});i+=j;continue;}
  if(c>='0'&&c<='9'){size_t j=1;while(j<r.size()&&r[j]>='0'&&r[j]<='9')++j;out.push_back({3,r.substr(0,j)});i+=j;continue;}
  size_t n=c<128?1:c<224?2:c<240?3:4;
  if(i+n>s.size())return false;out.push_back({4,r.substr(0,n)});i+=n;
 }
 out.push_back({5,{}});return true;
}
bool featureBoundary(const Expr&e,const std::vector<Token>&tokens,Match m){
 if(!m.ok||e.op!="seq"||e.args.size()!=2)return false;
 auto&a=expressions[e.args[0]];auto&b=expressions[e.args[1]];
 if(a.op!="lit"||a.text!=","||b.op!="assign"||b.text!="entityAttributes")return false;
 auto&t=tokens[m.end];return t.kind==2||(t.kind==0&&(t.text=="a"||t.text=="an"||t.text=="the"));
}
struct Key{std::string_view name;size_t pos;bool operator==(const Key&)const=default;};
struct Hash{size_t operator()(const Key&k)const{return std::hash<std::string_view>{}(k.name)^(k.pos*0x9e3779b97f4a7c15ULL);}};
struct Parser {
 const std::vector<Token>&tokens;
 std::unordered_map<Key,Match,Hash>memo;
 int depth=0;size_t steps=0;bool limited=false;
 Match rule(std::string_view name,size_t pos){
  int kind=name=="ID"?1:name=="STRING"?2:name=="INT"?3:-1;
  if(kind>=0)return tokens[pos].kind==kind?Match{pos+1,true}:Match{pos,false};
  if(name=="SL_COMMENT"||name=="ML_COMMENT")return{pos,false};
  Key key{name,pos};if(auto it=memo.find(key);it!=memo.end())return it->second;
  if(++depth>256){limited=true;--depth;return{pos,false};}
  auto it=rules.find(name);auto m=it==rules.end()?Match{pos,false}:eval(it->second,pos);
  --depth;memo.emplace(key,m);return m;
 }
 Match sequence(const std::vector<int>&args,size_t pos,size_t index){
  if(index==args.size())return{pos,true};auto&e=expressions[args[index]];auto m=eval(args[index],pos);
  if(m.ok){auto rest=sequence(args,m.end,index+1);if(rest.ok)return rest;}
  if(e.op=="repeat"&&e.text=="?")return sequence(args,pos,index+1);return{pos,false};
 }
 Match eval(int index,size_t pos){
  if(++steps>20000000){limited=true;return{pos,false};}auto&e=expressions[index];
  if(e.op=="lit")return tokens[pos].kind==0&&tokens[pos].text==e.text?Match{pos+1,true}:Match{pos,false};
  if(e.op=="call")return rule(e.text,pos);
  if(e.op=="ref")return rule("ID",pos);
  if(e.op=="action")return{pos,true};
  if(e.op=="assign")return eval(e.args[0],pos);
  if(e.op=="seq")return sequence(e.args,pos,0);
  if(e.op=="alt"){Match best{pos,false};for(int a:e.args){auto m=eval(a,pos);if(m.ok&&(!best.ok||m.end>best.end))best=m;}return best;}
  if(e.op=="repeat"){size_t at=pos,count=0;while(true){auto m=eval(e.args[0],at);if(featureBoundary(expressions[e.args[0]],tokens,m)||!m.ok||m.end==at)break;at=m.end;++count;if(e.text=="?")break;}return{at,e.text!="+"||count>0};}
  if(e.op=="unordered"){
   size_t at=pos;uint64_t used=0;if(e.args.size()>64){limited=true;return{pos,false};}
   while(true){int bestIndex=-1;Match best{at,false};for(size_t i=0;i<e.args.size();++i){if(used&(uint64_t(1)<<i))continue;auto m=eval(e.args[i],at);if(m.ok&&m.end>best.end){bestIndex=i;best=m;}}if(bestIndex<0)break;used|=uint64_t(1)<<bestIndex;at=best.end;}
   for(size_t i=0;i<e.args.size();++i)if(!(used&(uint64_t(1)<<i))){auto m=eval(e.args[i],at);if(!m.ok)return m;}return{at,true};
  }
  return{pos,false};
 }
};
bool check(std::string_view source){std::vector<Token>tokens;if(!lex(source,tokens))return false;Parser p{tokens,{}};auto m=p.rule("ContextMappingModel",0);return m.ok&&m.end==tokens.size()-1&&!p.limited;}
int main(int argc,char**argv){
 if(argc<2)return 2;std::ifstream f(argv[1]);if(!f)return 2;std::string s((std::istreambuf_iterator<char>(f)),{});
 size_t n=argc>2?std::stoull(argv[2]):200;if(!n)return 2;if(!check(s))return 1;
 auto begin=std::chrono::steady_clock::now();for(size_t i=0;i<n;++i)if(!check(s))return 1;
 auto ns=std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now()-begin).count()/n;
 std::cout<<"{\"language\":\"cpp\",\"bytes\":"<<s.size()<<",\"iterations\":"<<n<<",\"ns_per_op\":"<<ns<<",\"syntax_only\":true}\n";
}
