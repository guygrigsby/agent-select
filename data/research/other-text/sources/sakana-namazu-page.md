<!DOCTYPE html>
<html lang="ja" class="lang-ja">
  <head>
    <!-- Google Analytics (consistent with the rest of the site) -->
    <script async src="https://www.googletagmanager.com/gtag/js?id=G-W770C6CJ1J"></script>
    <script>
      window.dataLayer = window.dataLayer || [];
      function gtag(){dataLayer.push(arguments);}
      gtag('js', new Date());
      gtag('config', 'G-W770C6CJ1J');
    </script>

    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="description" content="Sakana Namazu is a Japanese-specialized LLM combining deep cultural understanding with high-performance reasoning, web search, and code execution.">
    <meta property="og:title" content="Sakana Namazu — 日本の感性で、考え、動くLLM">
    <meta property="og:type" content="website">
    <meta property="og:description" content="LLM with Japanese Vibes">
    <meta property="og:image" content="https://sakana.ai/assets/namazu-lp/ogp.png">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta name="twitter:image" content="https://sakana.ai/assets/namazu-lp/ogp.png">
    <meta property="og:url" content="https://sakana.ai/namazu/">
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:site" content="@SakanaAILabs">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&family=Noto+Sans+JP:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
    <title>Sakana Namazu — 日本の感性で、考え、動くLLM</title>
    <!-- 言語の初期化：チラつき防止のため描画前に localStorage を復元（既定は日本語）。
         切り替え自体はヘッダーの言語トグル（sakana-namazu-lp.js）で行う -->
    <script>
      (function () {
        try {
          var l = localStorage.getItem('namazu-lang');
          if (l !== 'ja' && l !== 'en') l = 'ja'; /* Japanese-first LP（EN版公開時に navigator 判定へ） */
          var r = document.documentElement;
          if (l === 'ja') { r.classList.add('lang-ja'); r.setAttribute('lang', 'ja'); }
          else { r.classList.remove('lang-ja'); r.setAttribute('lang', 'en'); }
        } catch (e) {}
      })();
    </script>
    <!-- ユースケースのDOM実装デモアニメ。スタイル・挙動とも uc-anim/ に分離。
         codex（旧case3・Codex再現デモ）は取り下げ中＝参照なし（ファイルは uc-anim/ に残置） -->
    <link rel="stylesheet" href="uc-anim/support.css">
    <link rel="stylesheet" href="uc-anim/research.css">
    <!-- デザイン確定版（docs/reference/design_handoff_namazu_lp）の自己完結スタイル。
         旧 assets/css/namazu-lp.css（fugu コピー）は廃止。プロトタイプの .fg-root.* スコープは
         html.* に読み替え済み。uc-anim の link より後に置き、デモ側の共通規則を確実に上書きする -->
    <style>


  *,*::before,*::after{box-sizing:border-box;}
  img,svg,video{display:block;}
  img,video{max-width:100%;height:auto;}
  ul,ol,menu{list-style:none;margin:0;padding:0;}
  figure,figcaption,blockquote{margin:0;padding:0;}
  button{font:inherit;color:inherit;background:transparent;border:0;cursor:pointer;}
  a{color:inherit;text-decoration:inherit;}
  .min-h-screen{min-height:100vh;}
  .bg-white{background:#fff;}
  .block{display:block;}
  .flex{display:flex;}
  .hidden{display:none;}
  .sr-only{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip-path:inset(50%);white-space:nowrap;border:0;}
  .t-ja{display:none;font-family:var(--font-jp);}
  html.lang-ja .t-en{display:none;}
  html.lang-ja .t-ja{display:revert;}
  [data-site-header][data-nav-collapsed="true"] [data-desktop-nav]{display:none!important;}
  [data-site-header][data-nav-collapsed="true"] [data-mobile-menu-button]{display:inline-flex!important;}
  [data-site-header][data-nav-collapsed="true"] [data-mobile-menu][data-open="true"]{display:block!important;}


			:root{
				--ac:#e10600;
				--font-en:"Poppins","Noto Sans JP",-apple-system,BlinkMacSystemFont,sans-serif;
				--font-jp:"Noto Sans JP","Hiragino Sans","Hiragino Kaku Gothic ProN","Yu Gothic",sans-serif;
				--font-mono-jb:"JetBrains Mono",ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;
			}
			.fg-root{font-family:var(--font-en);color:#111827;background:#fff;-webkit-font-smoothing:antialiased;overflow-x:clip;}
			::selection{background:rgba(225,6,0,.14);}
			.fg-mono{font-family:var(--font-mono-jb);}
			.fg-en{font-family:var(--font-en);}
			.fg-navlink .fg-en,.fg-aside__label .fg-en{font-family:var(--font-mono-jb);}
			/* Fugu LP と同一のレイアウトシステム：コンテンツ幅1280px・左レール200px。
			   モジュール(カード列・図版・FAQ・CTA)はすべて本文列の左端に揃える */
			.fg-wrap{max-width:1280px;margin:0 auto;}

			/* ---- header(Fugu LP と同構成) ---- */
			/* ヘッダーの内容箱をセクションと同じ構造(外側padding→内側1280px)にし、
			   ロゴと本文の左右端を全ビューポート幅で厳密に揃える(旧:内側paddingで広画面時に48pxズレ) */
			.fg-header{position:sticky;top:0;z-index:50;background:#fff;border-bottom:1px solid #e5e7eb;padding:0 clamp(20px,4vw,48px);transition:transform .25s ease,background .2s ease,box-shadow .2s ease;}
			.fg-header[data-scrolled="true"]{background:rgba(255,255,255,.7);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);box-shadow:0 8px 24px rgba(15,23,42,.06);}
			.fg-header[data-hidden="true"]{transform:translateY(-110%);}
			html.fg-menu-open .fg-header,html.fg-menu-open .fg-header[data-scrolled="true"]{backdrop-filter:none;-webkit-backdrop-filter:none;}
			.fg-header__inner{max-width:1280px;margin:0 auto;padding:0;min-height:clamp(64px,6vw,104px);display:flex;align-items:center;justify-content:space-between;gap:32px;}
			.fg-brand{display:flex;align-items:center;gap:10px;text-decoration:none;flex:none;}
			/* ロゴSVG(viewBox 3717×378、余白なし)。Fugu のロゴPNG は余白込み42px＝文字実寸21.9px のため、
			   文字実寸で揃える：ヘッダー26px・フッター28px(インク占有率の実測から換算) */
			.nz-logo{height:20px;width:auto;display:block;}
			.fg-brand__sub{display:none;font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.02em;color:#6b7280;white-space:nowrap;}
			/* JA側の文言も英字なのでmonoのまま(.t-jaの既定Noto Sans JPを打ち消す) */
			.fg-brand__sub .t-ja{font-family:var(--font-mono-jb);}
			@media(min-width:1360px){html.lang-ja .fg-brand__sub{display:inline;}}
			@media(min-width:1400px){html:not(.lang-ja) .fg-brand__sub{display:inline;}}
			.fg-actions{display:flex;align-items:center;gap:8px;min-width:0;}
			.fg-nav{display:flex;align-items:center;gap:6px;min-width:0;}
			.fg-nav ul{display:flex;align-items:center;gap:4px;list-style:none;margin:0;padding:0;overflow-x:auto;scrollbar-width:none;}
			.fg-nav ul::-webkit-scrollbar{display:none;}
			.fg-navlink{font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.02em;color:#374151;text-decoration:none;padding:8px;white-space:nowrap;background:transparent;border:0;cursor:pointer;transition:color .15s ease;}
			.fg-navlink:hover{color:var(--ac);}
			.fg-navlink[data-active="true"]{color:var(--ac);}
			.fg-pill{display:inline-flex;align-items:center;gap:8px;height:38px;padding:0 12px;background:#fff;border:1px solid #e5e7eb;border-radius:999px;cursor:pointer;font-family:var(--font-mono-jb);font-size:13px;font-weight:500;letter-spacing:.02em;color:#374151;white-space:nowrap;transition:border-color .15s ease,color .15s ease;}
			/* fuguはピルのJA側ラベルもmono強制だったが("Japanese"表記のため)、
			   「日本語」表記に変えたのでCJKが載るNoto Sans JP(.t-jaの既定)に任せる */
			.fg-pill:hover{border-color:var(--ac);color:var(--ac);}
			.fg-langwrap{position:relative;}
			.fg-pill__chev{color:#6b7280;transition:transform .15s ease;}
			[data-lang-toggle][aria-expanded="true"]{border-color:var(--ac);color:var(--ac);}
			[data-lang-toggle][aria-expanded="true"] .fg-pill__chev{transform:rotate(180deg);}
			.fg-langmenu{display:none;position:absolute;right:0;top:48px;min-width:160px;background:#fff;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 12px 32px rgba(15,23,42,.12);padding:8px;z-index:80;}
			.fg-langmenu[data-open="true"]{display:flex;flex-direction:column;}
			.fg-langopt{font-family:var(--font-jp);text-align:left;padding:10px 14px;background:transparent;border:0;border-radius:8px;cursor:pointer;font-size:15px;font-weight:600;color:#374151;}
			.fg-langopt:hover{background:#f9fafb;}
			html:not(.lang-ja) [data-lang-option="en"]{color:var(--ac);}
			html.lang-ja [data-lang-option="ja"]{color:var(--ac);}
			.fg-burger{flex:none;width:44px;height:44px;border-radius:10px;background:transparent;border:0;display:flex;align-items:center;justify-content:center;cursor:pointer;padding:0;}
			[data-desktop-nav]{display:flex;}
			[data-mobile-menu-button]{display:none;}
			[data-mobile-menu]{display:none;}
			[data-site-header][data-nav-collapsed="true"] [data-mobile-menu][data-open="true"]{display:flex!important;}
			.fg-mobilemenu{position:fixed;inset:0;z-index:200;flex-direction:column;width:auto;max-width:none;background:#fff;border:0;box-shadow:none;padding:0;overflow-y:auto;-webkit-overflow-scrolling:touch;}
			.fg-mobilemenu li{list-style:none;}
			.fg-mm__bar{position:sticky;top:0;z-index:1;flex:none;display:flex;align-items:center;justify-content:space-between;min-height:44px;padding:0 clamp(20px,4vw,48px);background:#fff;}
			.fg-mm__bar .nz-logo{height:18px;}
			.fg-mm__close{flex:none;width:44px;height:44px;border-radius:10px;background:transparent;border:0;display:flex;align-items:center;justify-content:center;cursor:pointer;padding:0;color:#111827;}
			.fg-mm__body{flex:1;padding:16px 20px calc(28px + env(safe-area-inset-bottom));}
			.fg-mobilemenu ul{margin:0;padding:0;display:flex;flex-direction:column;}
			.fg-mobilemenu ul button{display:flex;align-items:center;justify-content:space-between;gap:10px;width:100%;text-align:left;font-family:var(--font-en);font-size:19px;font-weight:600;color:#111827;text-decoration:none;padding:16px 4px;border:0;background:transparent;cursor:pointer;}
			html.lang-ja .fg-mobilemenu ul button{font-family:var(--font-jp);}
			.fg-mobilemenu ul button:hover,.fg-mobilemenu ul button[data-active="true"]{color:var(--ac);}
			.fg-mm__chev{flex:none;color:#9ca3af;}
			.fg-mm__cta{display:flex;align-items:center;justify-content:center;height:56px;margin-top:20px;background:#111827;color:#fff;font-family:var(--font-en);font-size:15px;font-weight:600;text-decoration:none;border-radius:10px;}
			.fg-mm__cta:hover{background:#1f2937;}
			.fg-mm__divider{height:1px;background:#e5e7eb;margin:28px 0;}
			.fg-mm__label{font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.08em;color:#9ca3af;margin:0 0 16px;}
			.fg-mm__products{display:flex;flex-direction:column;gap:18px;}
			.fg-mm__product{font-family:var(--font-en);font-size:19px;font-weight:600;color:#111827;text-decoration:none;}
			html.lang-ja .fg-mm__product{font-family:var(--font-jp);}
			.fg-mm__product:hover{color:var(--ac);}
			.fg-mm__social{display:flex;align-items:center;gap:18px;margin-top:8px;}
			.fg-mm__social a{width:40px;height:40px;display:flex;align-items:center;justify-content:center;color:#111827;}
			.fg-mm__social a:hover{color:var(--ac);}
			.fg-mm__social svg{width:24px;height:24px;fill:currentColor;}
			.fg-mm__langrow{display:flex;align-items:center;justify-content:flex-end;gap:8px;margin-top:36px;}
			.fg-mm__lang{background:transparent;border:0;cursor:pointer;padding:2px 4px;font-family:var(--font-mono-jb);font-size:15px;font-weight:600;color:#9ca3af;}
			.fg-mm__lang[data-lang-option="ja"]{font-family:var(--font-jp);}
			html:not(.lang-ja) .fg-mm__lang[data-lang-option="en"],html.lang-ja .fg-mm__lang[data-lang-option="ja"]{color:#111827;text-decoration:underline;text-underline-offset:4px;}
			.fg-mm__sep{color:#d1d5db;font-family:var(--font-mono-jb);}
			html.fg-menu-open{overflow:hidden;}
			html.fg-menu-open .fg-mcta{display:none!important;}
			/* モバイル固定の下部CTAバー(760px以下でのみ表示) */
			.fg-mcta{display:none;}
			.fg-mcta__btn{display:flex;align-items:center;justify-content:center;height:50px;background:#111827;color:#fff;font-family:var(--font-jp);font-size:15px;font-weight:600;text-decoration:none;border-radius:10px;}
			.fg-mcta__btn:hover{background:#1f2937;}

			/* ---- buttons(Fugu LP と同一) ---- */
			.fg-btn{display:inline-flex;align-items:center;height:46px;padding:0 22px;font-size:14px;font-weight:600;text-decoration:none;border:1px solid transparent;cursor:pointer;}
			.fg-btn--dark{background:#111827;color:#fff;}
			.fg-btn--dark:hover{background:#1f2937;}
			.fg-btn--ghost{background:#fff;color:#111827;border-color:#d1d5db;}
			.fg-btn--ghost:hover{border-color:#9ca3af;}

			/* ---- hero(Fugu LP と同構成：h1＋タグライン＋リード＋CTA＋注記／右に動画) ---- */
			.fg-hero{padding:clamp(56px,8vw,104px) clamp(20px,4vw,48px) 0;}
			.fg-hero__grid{max-width:1280px;margin:0 auto;display:flex;flex-wrap:wrap;gap:clamp(32px,5vw,72px);align-items:center;padding-bottom:clamp(48px,6vw,80px);}
			.fg-hero__col{flex:1 1 340px;min-width:260px;}
			.fg-hero__media{flex:1 1 300px;min-width:260px;}
			.fg-h1{margin:22px 0 0;font-family:var(--font-en);font-weight:500;font-size:clamp(42px,5.6vw,72px);line-height:1.02;letter-spacing:-.03em;color:#111827;}
			.fg-hero__cta{display:flex;gap:12px;flex-wrap:wrap;margin-top:30px;}
			.fg-hero__notice{margin:12px 0 0;max-width:520px;font-size:12.5px;line-height:1.7;color:#9ca3af;}
			.fg-herolead{font-size:16px;line-height:1.85;}
			/* ミニ動画は制作中のためプレースホルダー(完成後 <video> に差し替え) */
			.nz-placeholder{border:2px dashed #cbd5e1;background:#f8fafc;color:#94a3b8;display:flex;align-items:center;justify-content:center;text-align:center;font-size:13.5px;line-height:1.9;aspect-ratio:16/10;font-family:var(--font-jp);}
			/* hero のアイコンモーション：白背景の装飾動画なので枠なしでページに溶け込ませる */
			.nz-herowrap{position:relative;}
			.nz-herovideo{display:block;width:100%;max-width:520px;height:auto;margin:0 auto;}

			/* ---- 信頼帯(Fugu partners strip)。data-mode切替のJSコントローラは未導入：
			        ロゴ3枚なら既定の静的センター配置で足りるため。増えたらfuguから移植 ---- */
			.fg-partners{--fg-logos-gap:clamp(26px,4vw,46px);border-top:1px solid #e5e7eb;border-bottom:1px solid #e5e7eb;background:#fff;}
			.fg-partners__inner{max-width:1280px;margin:0 auto;padding:clamp(24px,3.4vw,34px) clamp(20px,4vw,48px);display:flex;align-items:center;justify-content:center;gap:clamp(28px,5vw,64px);flex-wrap:wrap;}
			.fg-partners__label{display:inline-flex;align-items:center;font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.14em;color:#6b7280;white-space:nowrap;flex:none;}
			/* ラベルは日英共通の英文なので、JAモードでもmonoのまま(fugu踏襲) */
			.fg-partners__label .t-ja{font-family:var(--font-mono-jb);}
			.fg-partners__vp{flex:0 1 auto;min-width:0;overflow:hidden;}
			.fg-partners__track{display:flex;align-items:center;justify-content:flex-start;width:max-content;max-width:100%;}
			.fg-partners__logos{display:flex;align-items:center;gap:var(--fg-logos-gap);flex:none;}
			.fg-partners__logo{display:flex;align-items:center;opacity:1;transition:opacity .15s ease;}
			.fg-partners__logo:hover{opacity:.55;}
			.fg-partners__logo img{width:auto;max-width:none;display:block;}
			.fg-partners__logo--or img{height:22px;}
			.fg-partners__logo--vc img{height:60px;margin:-18px 0;}
			@keyframes fugu-pmarquee{from{transform:translateX(0);}to{transform:translateX(-50%);}}
			@media (prefers-reduced-motion: reduce){.fg-partners__track{animation:none!important;}}

			/* ---- sections / rails(Fugu LP と同一：左に縦ラベル、右に本文のグリッド) ---- */
			.fg-sec{padding:clamp(64px,9vw,128px) clamp(20px,4vw,48px);}
			.fg-sec--alt{background:#fafbfc;}
			.fg-grid{max-width:1280px;margin:0 auto;display:grid;grid-template-columns:200px minmax(0,1fr);gap:clamp(28px,4vw,56px);}
			.fg-aside{position:sticky;top:120px;align-self:start;display:flex;flex-direction:column;align-items:flex-start;}
			.fg-aside__label{font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.14em;color:var(--ac);}
			.fg-aside__bar{width:28px;height:2px;background:var(--ac);margin-top:10px;}
			.fg-h2{margin:0;font-size:clamp(26px,3vw,36px);font-weight:700;line-height:1.4;color:#111827;}
			.fg-lead{margin:20px 0 0;font-size:16px;line-height:1.95;color:#6b7280;max-width:680px;}
			.fg-link{color:var(--ac);text-decoration:underline;text-underline-offset:2px;}

			/* ---- cells(とは の3ポイント。Fugu benefits と同一) ---- */
			.fg-cells{display:grid;gap:1px;background:#e5e7eb;border:1px solid #e5e7eb;margin-top:clamp(32px,4vw,48px);}
			/* 3列が入らない幅では2×2にせず1×3の縦積みに(空セルが生まれるため) */
			.fg-cells--3{grid-template-columns:repeat(3,minmax(0,1fr));}
			@media(max-width:1019px){.fg-cells--3{grid-template-columns:minmax(0,1fr);}}
			.fg-cell{background:#fff;padding:26px 24px;}
			.fg-cell__n{font-family:var(--font-mono-jb);font-size:12px;color:var(--ac);letter-spacing:.06em;}
			.fg-cell__t{font-size:16px;font-weight:600;color:#111827;margin-top:14px;line-height:1.5;}
			.fg-cell__b{margin:10px 0 0;font-size:13.5px;line-height:1.85;color:#6b7280;}

			/* モジュール下のボタン行(Fugu 同様、列の左端に揃える) */
			.nz-secta{margin-top:clamp(28px,3.5vw,40px);display:flex;justify-content:flex-start;gap:12px;flex-wrap:wrap;}
			/* モバイル・タブレット幅ではCTAをFill(行全幅を等分して横並び)にする */
			@media(max-width:1019px){
				.nz-secta .fg-btn,.nz-qs__cta .fg-btn,.fg-hero__cta .fg-btn,.nz-contact-cta .fg-btn{flex:1 1 0;min-width:0;justify-content:center;}
			}

			/* ---- quick start：コードブロック(最終コード＋変更行ハイライト、
			        上部は言語タブ curl / Python / JavaScript) ---- */
			.nz-codewrap{margin-top:clamp(32px,4vw,48px);border:1px solid #e5e7eb;background:#fff;overflow:hidden;}
			.nz-codetabs{display:flex;gap:4px;background:#f6f8fa;border-bottom:1px solid #e5e7eb;padding:8px 10px 0;}
			.nz-codetab{font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.02em;color:#6b7280;background:transparent;border:1px solid transparent;border-bottom:0;padding:7px 14px;cursor:pointer;}
			/* アクティブタブ：下のコード面(白)と一体に見せるため境界線に1px重ねる */
			.nz-codetab[data-active="true"]{background:#fff;border-color:#e5e7eb;color:#111827;margin-bottom:-1px;}
			.nz-code{margin:0;background:#fff;color:#1f2937;font-family:var(--font-mono-jb);font-size:13.5px;line-height:1.85;padding:22px 26px;overflow-x:auto;}
			/* .chg = OpenAI からの差し替え行（base_url / api_key / model）。HTML 側に改行が無いため
			   display:block が行構造を担う（span 除去や inline 化は行が潰れるので不可）。
			   強調のあしらい（淡赤背景+左バー）は削除行に見えるため廃止＝素のコード表示 */
			.nz-code .chg{display:block;}
			/* 言語タブのパネルは同じグリッドマスに重ねる：枠の高さ＝常に両タブの最大値になり、
			   タブ・言語の切替で下のコンテンツが上下に揺れない。非表示側は visibility で不可視化 */
			/* コード面の右上に浮かせる（横スクロールしても定位置）。コードの上に載るため白地＋枠で浮かす */
			.nz-codecopy{position:absolute;top:10px;right:10px;z-index:1;display:flex;align-items:center;justify-content:center;width:30px;height:30px;color:#6b7280;background:#fff;border:1px solid #e5e7eb;border-radius:6px;padding:0;cursor:pointer;box-shadow:0 1px 2px rgba(0,0,0,.06);transition:color .15s ease,background .15s ease,border-color .15s ease;}
			.nz-codecopy:hover{color:#111827;border-color:#9ca3af;}
			.nz-codecopy svg{display:block;}
			.nz-codecopy .ic-check{display:none;color:#1a7f37;}
			.nz-codecopy[data-copied] .ic-copy{display:none;}
			.nz-codecopy[data-copied] .ic-check{display:block;}
			.nz-codepanes{display:grid;position:relative;}
			.nz-codepanes .nz-code{grid-area:1/1;min-width:0;}
			.nz-codepanes .nz-code[data-active="false"]{visibility:hidden;}
			/* quick start 2カラム：左＝見出し・手順・CTA、右＝diffコード(列内・左揃え) */
			.nz-qs{display:flex;flex-wrap:wrap;gap:clamp(32px,4vw,64px);align-items:center;}
			.nz-qs__col{flex:1 1 340px;min-width:280px;}
			.nz-qs__code{flex:1 1 440px;min-width:300px;margin-top:0;}
			.nz-qs__coderail{flex:1 1 440px;min-width:300px;}
			.nz-qs__coderail .nz-qs__cta{margin-top:18px;}
			/* CTAの位置：広幅=手順の下(--wide)のみ、スマホ幅=コードdiffの下(--narrow)のみ。
			   下の .nz-qs__cta{display:flex} に負けないよう2クラスで詳細度を上げる */
			.nz-qs__cta.nz-qs__cta--narrow{display:none;}
			@media(max-width:560px){
				.nz-qs__cta.nz-qs__cta--wide{display:none;}
				.nz-qs__cta.nz-qs__cta--narrow{display:flex;}
			}
			.nz-qs__steps{list-style:none;margin:30px 0 0;padding:0;display:flex;flex-direction:column;gap:20px;}
			.nz-qs__step{display:grid;grid-template-columns:34px 1fr;gap:12px;}
			.nz-qs__step .n{font-family:var(--font-mono-jb);font-size:13px;color:var(--ac);letter-spacing:.06em;padding-top:2px;}
			.nz-qs__step h3{margin:0;font-size:15px;font-weight:600;color:#111827;line-height:1.5;}
			.nz-qs__step p{margin:4px 0 0;font-size:13.5px;line-height:1.8;color:#6b7280;}
			.nz-qs__cta{display:flex;gap:12px;flex-wrap:wrap;margin-top:32px;}

			/* ---- models / pricing(fuguのplan・pricegridコンポーネント準拠) ---- */
			.fg-plan{border:1px solid #e5e7eb;background:#fff;padding:clamp(22px,3vw,32px);margin-top:clamp(32px,4vw,48px);}
			.fg-plan__head{display:flex;justify-content:space-between;align-items:baseline;gap:12px;padding-bottom:16px;border-bottom:1px solid #e5e7eb;}
			.fg-plan__title{font-size:18px;font-weight:700;color:#111827;margin-left:8px;}
			.fg-tag{font-family:var(--font-mono-jb);font-size:11px;padding:4px 10px;border:1px solid #111827;color:#111827;white-space:nowrap;}
			.fg-tag--ac{border-color:var(--ac);color:var(--ac);}
			.fg-pricegrid{display:grid;grid-template-columns:repeat(3,1fr);gap:1px;background:#e5e7eb;border:1px solid #e5e7eb;margin-top:20px;}
			.nz-pricegrid--2{grid-template-columns:repeat(2,1fr);}
			.fg-pricecell{background:#fff;padding:18px 14px;}
			.fg-pricecell__l{font-size:11px;color:#6b7280;}
			.fg-pricecell__p{font-family:var(--font-mono-jb);font-size:30px;font-weight:600;line-height:1.1;color:#111827;margin-top:8px;}
			.fg-pricecell__n{font-size:10px;line-height:1.5;color:#9ca3af;margin-top:10px;}
			.fg-pricecell__jpy{font-family:var(--font-mono-jb);font-size:13px;color:#6b7280;margin-top:6px;}
			.nz-toolshead{margin:24px 0 0;font-size:14px;font-weight:700;color:#111827;}
			.nz-note{margin:24px 0 0;font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.02em;line-height:1.9;color:#9ca3af;}

			/* ---- capability：ベンチマーク図・デモ(フル幅1行1枚) ---- */
			.nz-fig{margin-top:clamp(32px,4vw,48px);border:1px solid #e5e7eb;background:#fff;position:relative;}
			.nz-fig__pad{padding:clamp(16px,2.5vw,32px);}
			.nz-fig__pad img{width:100%;max-width:880px;display:block;margin:0 auto;}
			.nz-figcap{font-size:13px;line-height:1.8;color:#6b7280;border-top:1px solid #e5e7eb;padding:12px 16px;}

			/* ---- 利用実績 / Case Study（1件フィーチャー版・プロトタイプ） ---- */
			.nz-cs{margin-top:clamp(32px,4vw,48px);border:1px solid #e5e7eb;background:#fff;display:flex;flex-wrap:wrap;}
			.nz-cs__main{flex:1 1 100%;min-width:280px;padding:clamp(24px,3vw,40px);}
			.nz-cs__b,.nz-cs__t{max-width:none;}
			.nz-cs__logo{display:flex;align-items:center;gap:14px;}
			.nz-cs__logo img,.nz-cs__logo svg{height:30px;width:auto;display:block;}
			.nz-cs__co{font-size:14px;font-weight:600;letter-spacing:.02em;color:#111827;padding-left:14px;border-left:1px solid #e5e7eb;}
			.nz-cs__meta{margin-top:14px;font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.06em;color:var(--ac);border-bottom:1px solid var(--ac);padding-bottom:2px;display:inline-flex;align-self:flex-start;}
			.nz-cs__t{margin:14px 0 0;font-size:clamp(18px,2vw,22px);font-weight:600;line-height:1.55;color:#111827;}
			.nz-cs__b{margin:14px 0 0;font-size:14px;line-height:1.9;color:#6b7280;}
			.nz-cs__flow{margin-top:22px;display:flex;flex-wrap:wrap;gap:8px;}
			.nz-cs__step{font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.04em;color:#374151;border:1px solid #e5e7eb;padding:6px 10px;}
			.nz-cs__stats{margin-top:18px;display:flex;flex-wrap:wrap;gap:0;border:1px solid #e5e7eb;}
			.nz-cs__stat{flex:1 1 180px;padding:16px 20px;border-right:1px solid #e5e7eb;}
			.nz-cs__stat:last-child{border-right:0;}
			@media (max-width:640px){.nz-cs__stat{border-right:0;border-bottom:1px solid #e5e7eb;}.nz-cs__stat:last-child{border-bottom:0;}}
			.nz-cs__num{font-family:var(--font-mono-jb);font-size:26px;font-weight:600;color:#111827;line-height:1.1;}
			.nz-cs__cap{margin-top:5px;font-size:11.5px;line-height:1.6;letter-spacing:.02em;color:#9ca3af;}
			.nz-cs__fn{margin:16px 0 0;font-size:11.5px;line-height:1.9;color:#9ca3af;}
			.nz-cs__note{margin-top:14px;font-size:12px;color:#9ca3af;}
			.nz-contact-nowrap{white-space:nowrap;}


			/* ---- use cases(fugu foundationsの行形式：1行1件、本文＋メディア) ---- */
			.nz-ucs{margin-top:clamp(32px,4vw,48px);border:1px solid #e5e7eb;background:#fff;}
			.nz-uc{position:relative;display:flex;flex-wrap:wrap;gap:clamp(20px,3vw,40px);align-items:center;padding:clamp(20px,3vw,32px);border-bottom:1px solid #e5e7eb;}
			.nz-uc:last-child{border-bottom:0;}
			.nz-uc__body{flex:1 1 340px;min-width:260px;}
			.nz-uc__media{flex:1 1 320px;min-width:240px;}
			.nz-uc__tag{display:inline-flex;align-items:center;gap:6px;font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.06em;color:var(--ac);border-bottom:1px solid var(--ac);padding-bottom:2px;}
			.nz-uc__t{margin:10px 0 0;font-size:18px;font-weight:600;line-height:1.5;color:#111827;}
			.nz-uc__b{margin:12px 0 0;font-size:13.5px;line-height:1.85;color:#6b7280;}
			/* 強調は Fugu と同じ扱い：地色なし・濃色の太字 */
			.nz-hl{font-weight:700;color:#0f172a;}
			.nz-chips{display:flex;flex-wrap:wrap;gap:8px;margin-top:14px;}
			.nz-chip{font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.02em;border:1px solid #e5e7eb;background:#fafbfc;color:#6b7280;padding:3px 10px;}
			/* Research PJとの掲載調整が未了であることを示す社内向けバッジ */
			.nz-adjbadge{position:absolute;top:12px;right:12px;background:#eef2ff;color:#3538cd;border:1px dashed #a5b4fc;border-radius:6px;font-size:11px;padding:2px 10px;font-weight:600;z-index:1;}

			/* ---- faq(Fugu LP と同一のアコーディオン。開閉は sakana-namazu-lp.js §4) ---- */
			.fg-faqs{border-top:1px solid #111827;}
			.fg-faq{border-bottom:1px solid #e5e7eb;}
			.fg-faq__btn{width:100%;display:grid;grid-template-columns:48px 1fr 28px;gap:16px;align-items:center;padding:22px 4px;background:transparent;border:0;cursor:pointer;text-align:left;}
			.fg-faq__n{font-family:var(--font-mono-jb);font-size:13px;color:var(--ac);}
			.fg-faq__q{font-size:clamp(15px,1.4vw,17px);font-weight:600;color:#111827;line-height:1.5;}
			.fg-faq__icon{color:var(--ac);display:flex;justify-content:center;}
			.fg-faq__icon .fg-minus{display:none;}
			[data-faq-item][data-open="true"] .fg-faq__icon .fg-minus{display:block;}
			[data-faq-item][data-open="true"] .fg-faq__icon .fg-plus{display:none;}
			.fg-faq__ans{display:grid;grid-template-rows:0fr;transition:grid-template-rows .22s ease;}
			[data-faq-answer][data-open="true"]{grid-template-rows:1fr;}
			.fg-faq__ans p{margin:0;min-height:0;overflow:hidden;padding:0 4px 0 64px;font-size:14px;line-height:1.9;color:#6b7280;transition:padding-bottom .22s ease;}
			[data-faq-answer] p[data-open="true"]{padding-bottom:24px;}
			/* 回答内の社内向け注記(要確認・法務審査待ち等) */
			.nz-qnote{display:block;margin-top:4px;font-size:12px;color:#8a5a00;}

			/* ---- footer(Fugu LP と同構成：左下ロゴ＋右リンク列＋mono copyright) ---- */
			.fg-footer{background:#fff;padding:48px clamp(20px,4vw,48px);border-top:1px solid #e5e7eb;}
			.fg-footer__inner{max-width:1280px;margin:0 auto;display:grid;grid-template-columns:1fr minmax(0,500px);grid-template-rows:1fr auto;grid-template-areas:"logo links" "copy links";column-gap:64px;row-gap:24px;min-height:280px;}
			.fg-footer__logo{grid-area:logo;align-self:end;}
			/* FuguのロゴPNGはグリフ下に透過余白があり実質gapが広い。余白なしSVGでは
			   marginで補正し、コピーライトとの実質gapを 24+12=36px(4の倍数)に揃える */
			.fg-footer__logo .nz-logo{height:22px;margin-bottom:12px;}
			.fg-footer__links{grid-area:links;align-self:start;display:flex;flex-direction:column;}
			.fg-footer__links a{font-family:var(--font-jp);display:flex;align-items:center;justify-content:space-between;gap:12px;padding:20px 0;border-bottom:1px solid #e5e7eb;text-decoration:none;color:#111827;font-size:15px;transition:color .15s ease;}
			.fg-footer__links a:hover{color:var(--ac);}
			.fg-footer__links a svg{flex:none;stroke:#9ca3af;transition:stroke .15s ease;}
			.fg-footer__links a:hover svg{stroke:var(--ac);}
			.fg-footer__copy{grid-area:copy;align-self:end;font-family:var(--font-mono-jb);font-size:12px;color:#9ca3af;}

			/* ---- responsive(セクション追加のたびに拡張) ---- */
			@media(max-width:760px){
				.fg-header__inner{min-height:44px;}
				.nz-logo{height:18px;}
				.fg-langwrap{display:none;}
				.fg-sec{padding:48px clamp(16px,4vw,20px);}
				.fg-grid{grid-template-columns:minmax(0,1fr);gap:22px;}
				.fg-aside{position:static;top:auto;margin-bottom:16px;}
				.fg-pricegrid,.nz-pricegrid--2{grid-template-columns:1fr;}
				.fg-faq__btn{grid-template-columns:32px 1fr 24px;gap:10px;padding:18px 2px;}
				.fg-faq__ans p{padding-left:44px;}
				.fg-footer{padding-bottom:96px;}
				.fg-footer__inner{grid-template-columns:1fr;grid-template-rows:auto auto auto;grid-template-areas:"logo" "links" "copy";min-height:0;row-gap:0;}
				.fg-footer__links{margin-top:36px;}
				.fg-footer__copy{margin-top:36px;}
				/* hero：1カラム積み(テキスト→動画)。インラインCTAは隠し、下部固定バーが主CTAを担う */
				.fg-hero{padding:24px clamp(16px,4vw,20px) 0;}
				.fg-hero__grid{flex-direction:column;align-items:stretch;gap:clamp(24px,5vw,40px);}
				.fg-hero__col,.fg-hero__media{flex:1 1 auto;min-width:0;width:100%;}
				.fg-hero__cta{display:none;}
				.fg-h1{font-size:clamp(40px,12vw,52px);}
				.fg-herolead{font-size:13.5px;line-height:1.6;}
				.fg-herosub{display:none;}
				.fg-mcta{display:block;position:fixed;left:0;right:0;bottom:0;z-index:55;background:rgba(255,255,255,.95);backdrop-filter:blur(10px);-webkit-backdrop-filter:blur(10px);border-top:1px solid #e5e7eb;padding:14px 16px calc(14px + env(safe-area-inset-bottom));}
				.fg-partners{--fg-logos-gap:30px;}
				.fg-partners__inner{padding-top:30px;padding-bottom:30px;}
			}

			/* ---- use-case demo shells（再生オーバーレイ・全幅ステージ・オンボーディング式ツールチップ） ---- */

			/* ユースケースデモの中央再生ボタン(初回はクリックで再生開始)。3ケース共通 */
			.nz-anim1__in,.nz-anim2__in,.nz-anim3__in{position:relative;}
			.nz-uc__play{position:absolute;inset:0;z-index:30;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:14px;width:100%;height:100%;border:0;cursor:pointer;background:rgba(255,255,255,.10);-webkit-backdrop-filter:blur(3px) saturate(160%);backdrop-filter:blur(3px) saturate(160%);color:#111827;font-family:var(--font-jp);}
			.nz-uc__play-ic{display:flex;align-items:center;justify-content:center;width:64px;height:64px;border-radius:50%;background:var(--ac);color:#fff;font-size:24px;padding-left:4px;box-shadow:0px 4px 6px -1px rgba(0,0,0,.1),0px 2px 4px -2px rgba(0,0,0,.1);transition:transform .15s ease;}
			.nz-uc__play:hover .nz-uc__play-ic,.nz-uc__play:focus-visible .nz-uc__play-ic{transform:scale(1.08);}
			.nz-uc__play-tx{font-size:13px;font-weight:600;letter-spacing:.02em;}
			.nz-uc__play[hidden]{display:none;}
			/* ユースケース：デモ行を全幅ブレークアウトし、左右に余白カラムを確保（レール幅200pxに追従） */
			.nz-uc--wide .nz-uc__media{max-width:none;}
			.nz-uc__stage{width:calc(100% + 200px + clamp(28px,4vw,56px));margin-left:calc(-200px - clamp(28px,4vw,56px));transform:translateX(calc((200px + clamp(28px,4vw,56px)) / 2));display:grid;grid-template-columns:minmax(180px,1fr) minmax(0,820px) minmax(180px,1fr);gap:18px;align-items:stretch;}
			.nz-uc__demoslot{position:relative;min-width:0;}
			/* デモカード→注記のgap：Fuguのfg-benchcap(16px)。幅はステージ中央列(デモカード)の820pxに一致させる */
			.nz-anim1__note,.nz-anim2__note,.nz-anim3__note{max-width:820px;margin:16px auto 0;}
			/* CASE 03 は uc-anim/codex.css を読み込まないため、注記の字面スタイルをここで定義 */
			.nz-anim3__note{font-size:10.5px;line-height:1.7;color:#9ca3af;font-family:var(--font-jp);}
			/* CASE 03 の動画プレーヤー：デモカードと同じ枠・影（shadow-sm）で揃える。
			   position:relative は中央▶オーバーレイ（.nz-uc__play）の基準 */
			/* 幅は CASE 01/02 のデモ窓（stage 中央列の 820px）に揃える */
			.nz-videobox{position:relative;max-width:820px;margin:0 auto;border:1px solid #e5e7eb;border-radius:12px;overflow:hidden;background:#fff;box-shadow:0px 1px 3px 0px rgba(0,0,0,.1),0px 1px 2px -1px rgba(0,0,0,.1);}
			.nz-videobox video{display:block;width:100%;height:auto;}
			.nz-uc__b.nz-uc__b--wide{max-width:none;margin-top:24px;}
			.nz-uc__b--wide+.nz-uc__b--wide{margin-top:10px;}
			/* ツールチップ（オンボーディング式）：ビート到達でデモを一時停止し、説明対象だけを
			   スポットライトで残して他を暗転、白カードを対象の近くに重ねる。暗幕は SVG マスク方式
			   （黒 rect＝穴を複数開けられる。box-shadow 方式は穴が1つしか作れない）。
			   svg の生成と穴・カードの配置は caseN.js（initSpot / showTip）が行う */
			.nz-uc__spot{position:absolute;inset:0;z-index:28;opacity:0;transition:opacity .3s ease;pointer-events:none;}
			/* 表示中は暗幕がデモへのクリックを全て吸う：穴は凍結時の固定矩形のため、下で画面が
			   切り替わると高亮と実画面がズレる。クリックはチップを早めに閉じるだけ（caseN.js） */
			.nz-uc__spot.is-on{opacity:1;pointer-events:auto;}
			.nz-uc__spot svg{position:absolute;inset:0;width:100%;height:100%;display:block;}
			/* 再生が終わって ↻ が出るまで、デモの中身は操作不可（クリック・内部スクロールとも）：
			   再生途中に画面を切り替えられると、ツールチップの固定矩形の穴と実画面がズレるため。
			   chrome/topbar の ⏸/▶ は対象外で生かす */
			.nz-anim1--lock .nz-anim1__app,
			.nz-anim2--lock .nz-anim2__desk,
			.nz-anim3--lock .nz-anim3__stage{pointer-events:none;}
			/* カードは白地：暗幕の上では黒カードが沈むため */
			.nz-co{position:absolute;left:0;top:0;z-index:29;width:min(320px,calc(100% - 20px));background:#fff;color:#374151;border:1px solid #e5e7eb;border-radius:12px;padding:11px 13px;box-shadow:0 10px 26px rgba(15,23,42,.25);font-size:12.5px;line-height:1.7;opacity:0;transform:translateY(6px) scale(.97);transition:opacity .28s ease,transform .28s ease;pointer-events:none;}
			.nz-co.is-show{opacity:1;transform:none;}
			.nz-co .n{display:inline-flex;align-items:center;justify-content:center;width:18px;height:18px;border-radius:50%;background:var(--ac);color:#fff;font-family:var(--font-mono-jb);font-size:11px;font-weight:700;margin-right:6px;vertical-align:1px;}
			.nz-co .task{font-size:13px;color:#111827;font-weight:600;}
			.nz-co .cap{display:block;margin-top:5px;font-size:11px;color:#6b7280;font-weight:500;}
			.nz-co .cap b{color:var(--ac);}
			.nz-co .chip{display:inline-block;font-family:var(--font-mono-jb);font-size:10px;letter-spacing:.04em;background:#f3f4f6;color:#4b5563;border-radius:999px;padding:1px 8px;margin-top:8px;}
			/* 矢印はスポットの穴（説明対象）を指す。--ax/--ay は showTip が穴の中心に合わせて設定 */
			.nz-co::after{content:"";position:absolute;display:none;width:12px;height:12px;background:#fff;transform:rotate(45deg);left:var(--ax,24px);}
			.nz-co--b::after{display:block;top:-5px;box-shadow:-1px -1px 0 #e5e7eb;}
			.nz-co--a::after{display:block;bottom:-5px;box-shadow:1px 1px 0 #e5e7eb;}
			.nz-co--r::after{display:block;left:-5px;top:var(--ay,24px);box-shadow:-1px 1px 0 #e5e7eb;}
			.nz-co--l::after{display:block;left:auto;right:-5px;top:var(--ay,24px);box-shadow:1px -1px 0 #e5e7eb;}
			/* ツールチップ（スポットライト＋カード）は全幅で有効：旧・ガター常駐カードは
			   狭幅で映えず非表示にしていたが、現方式はデモ上のオーバーレイなので狭幅でも成立する */
			@media(max-width:1000px){
				.nz-uc__stage{display:flex;flex-direction:column;width:auto;margin-left:0;transform:none;gap:0;}
			}
    </style>
  </head>
  <body>
<div class="min-h-screen bg-white fg-root">
      <header data-site-header="" data-nav-collapsed="true" class="fg-header">
        <div data-header-inner="" class="fg-header__inner">
          <a href="#top" data-brand="" data-scroll-target="top" class="fg-brand">
            <img class="nz-logo" src="/assets/namazu-lp/logo-sakana-namazu-colored.svg" alt="Sakana Namazu">
            <span class="fg-brand__sub">
              <span class="t-en">japan-vibes LLM</span>
              <span class="t-ja">japan-vibes LLM</span>
            </span>
          </a>
          <div data-header-actions="" class="fg-actions">
            <nav data-desktop-nav="" class="fg-nav" aria-label="Primary">
              <ul>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="what-is" class="fg-navlink" aria-current="page" data-active="true">
                    <span class="t-en">
                      What is
                      <span class="fg-en">Sakana Namazu</span>
                      ?
                    </span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Namazu</span>
                      とは
                    </span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="quickstart" class="fg-navlink" data-active="false">
                    <span class="t-en">Quick Start</span>
                    <span class="t-ja">使い方</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="models" class="fg-navlink" data-active="false">
                    <span class="t-en">Pricing</span>
                    <span class="t-ja">料金</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="capability" class="fg-navlink" data-active="false">
                    <span class="t-en">Capabilities</span>
                    <span class="t-ja">モデル性能</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="works" class="fg-navlink" data-active="false">
                    <span class="t-en">Case Study</span>
                    <span class="t-ja">導入事例</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="use-cases" class="fg-navlink" data-active="false">
                    <span class="t-en">Use Cases</span>
                    <span class="t-ja">ユースケース</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="faq" class="fg-navlink" data-active="false">FAQ</button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="contact" class="fg-navlink" data-active="false">
                    <span class="t-en">Contact Us</span>
                    <span class="t-ja">お問い合わせ</span>
                  </button>
                </li>
              </ul>
            </nav>
            <div data-lang-switch="" class="fg-langwrap">
              <button type="button" data-lang-toggle="" data-lang="ja" class="fg-pill" aria-haspopup="true" aria-expanded="false" aria-controls="lang-menu" aria-label="Select language" title="Select language">
                <span class="sr-only">Select language</span>
                <span class="fg-pill__label">
                  <span class="t-en">English</span>
                  <span class="t-ja">日本語</span>
                </span>
                <svg class="fg-pill__chev" width="14" height="14" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="m6 9 6 6 6-6"></path>
                </svg>
              </button>
              <div id="lang-menu" data-lang-menu="" data-open="false" class="fg-langmenu" role="menu" aria-label="Select language">
                <button type="button" data-lang-option="ja" role="menuitemradio" aria-checked="true" class="fg-langopt">日本語</button>
                <button type="button" data-lang-option="en" role="menuitemradio" aria-checked="false" class="fg-langopt">English</button>
              </div>
            </div>
            <button type="button" data-mobile-menu-button="" class="fg-burger" aria-label="Open navigation menu" aria-controls="mobile-nav-menu" aria-expanded="false">
              <span class="sr-only">Open navigation menu</span>
              <svg width="26" height="26" viewbox="0 0 24 24" fill="none" stroke="#111827" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                <path d="M4 6h16"></path>
                <path d="M4 12h16"></path>
                <path d="M4 18h16"></path>
              </svg>
            </button>
          </div>
          <nav id="mobile-nav-menu" data-mobile-menu="" data-open="false" class="fg-mobilemenu" aria-label="Mobile navigation">
            <div class="fg-mm__bar">
              <img class="nz-logo" src="/assets/namazu-lp/logo-sakana-namazu-colored.svg" alt="Sakana Namazu">
              <button type="button" data-mobile-menu-close="" class="fg-mm__close" aria-label="Close navigation menu">
                <svg width="26" height="26" viewbox="0 0 24 24" fill="none" stroke="#111827" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                  <path d="M6 6l12 12"></path>
                  <path d="M18 6 6 18"></path>
                </svg>
              </button>
            </div>
            <div class="fg-mm__body">
              <ul>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="what-is" aria-current="page" data-active="true">
                    <span>
                      <span class="t-en">
                        What is
                        <span class="fg-en">Sakana Namazu</span>
                        ?
                      </span>
                      <span class="t-ja">
                        <span class="fg-en">Sakana Namazu</span>
                        とは
                      </span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="quickstart" data-active="false">
                    <span>
                      <span class="t-en">Quick Start</span>
                      <span class="t-ja">使い方</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="models" data-active="false">
                    <span>
                      <span class="t-en">Pricing</span>
                      <span class="t-ja">料金</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="capability" data-active="false">
                    <span>
                      <span class="t-en">Capabilities</span>
                      <span class="t-ja">モデル性能</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="use-cases" data-active="false">
                    <span>
                      <span class="t-en">Use Cases</span>
                      <span class="t-ja">ユースケース</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="faq" data-active="false">
                    <span>
                      <span class="t-en">FAQ</span>
                      <span class="t-ja">FAQ</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="contact" data-active="false">
                    <span>
                      <span class="t-en">Contact Us</span>
                      <span class="t-ja">お問い合わせ</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </li>
              </ul>
              <div class="fg-mm__divider"></div>
              <p class="fg-mm__label">Other Products</p>
              <div class="fg-mm__products">
                <a class="fg-mm__product" href="https://sakana.ai/marlin/" target="_blank" rel="noreferrer">Sakana Marlin</a>
                <a class="fg-mm__product" href="https://sakana.ai/fugu/" target="_blank" rel="noreferrer">Sakana Fugu</a>
              </div>
              <p class="fg-mm__label" style="margin-top:32px;">Follow us</p>
              <div class="fg-mm__social">
                <a href="https://x.com/SakanaAILabs" aria-label="X" target="_blank" rel="noopener">
                  <svg viewbox="0 0 24 24" aria-hidden="true">
                    <path d="M22.46 6c-.77.35-1.6.58-2.46.69.88-.53 1.56-1.37 1.88-2.38-.83.5-1.75.85-2.72 1.05C18.37 4.5 17.26 4 16 4c-2.35 0-4.27 1.92-4.27 4.29 0 .34.04.67.11.98C8.28 9.09 5.11 7.38 3 4.79c-.37.63-.58 1.37-.58 2.15 0 1.49.75 2.81 1.91 3.58-.71 0-1.37-.2-1.95-.5v.03c0 2.08 1.48 3.82 3.44 4.21a4.22 4.22 0 0 1-1.93.07 4.28 4.28 0 0 0 4 2.98 8.521 8.521 0 0 1-5.33 1.84c-.34 0-.68-.02-1.02-.06C3.44 20.29 5.7 21 8.12 21 16 21 20.33 14.46 20.33 8.79c0-.19 0-.37-.01-.56.84-.6 1.56-1.36 2.14-2.23z"></path>
                  </svg>
                </a>
                <a href="https://jp.linkedin.com/company/sakana-ai" aria-label="LinkedIn" target="_blank" rel="noopener">
                  <svg viewbox="0 0 24 24" aria-hidden="true">
                    <path d="M20.45 20.45h-3.55v-5.57c0-1.33-.02-3.04-1.85-3.04-1.85 0-2.14 1.45-2.14 2.94v5.67H9.36V9h3.41v1.56h.05c.48-.9 1.64-1.85 3.37-1.85 3.6 0 4.27 2.37 4.27 5.45v6.29zM5.34 7.43a2.06 2.06 0 1 1 0-4.12 2.06 2.06 0 0 1 0 4.12zM7.12 20.45H3.56V9h3.56v11.45zM22.22 0H1.77C.79 0 0 .77 0 1.72v20.56C0 23.23.79 24 1.77 24h20.45c.98 0 1.78-.77 1.78-1.72V1.72C24 .77 23.2 0 22.22 0z"></path>
                  </svg>
                </a>
              </div>
              <div class="fg-mm__langrow">
                <button type="button" data-lang-option="ja" class="fg-mm__lang">日本語</button>
                <span class="fg-mm__sep" aria-hidden="true">/</span>
                <button type="button" data-lang-option="en" class="fg-mm__lang">English</button>
              </div>
            </div>
          </nav>
        </div>
      </header>
      <main>
        <!-- §00 hero -->
        <section id="top" class="fg-hero">
          <div class="fg-hero__grid">
            <div class="fg-hero__col">
              <h1 class="fg-h1">
                <span class="fg-en">Sakana Namazu</span>
              </h1>
              <div class="fg-herosub" style="margin-top:16px;font-size:clamp(16px,1.8vw,19px);font-weight:600;color:#0f172a;">
                <span class="t-en">LLM with Japanese Vibes</span>
                <span class="t-ja">日本の感性で、考え、動くLLM</span>
              </div>
              <p class="fg-herolead" style="margin:20px 0 0;color:#6b7280;max-width:520px;">
                <span class="t-en"><span class="fg-en">Sakana Namazu</span> is an LLM that combines a deep understanding of Japanese culture with high-performance reasoning capabilities. Combining web search and code execution, it handles complex tasks across a wide range of business scenarios. Switching from your current model is easy. Please give it a try.</span>
                <span class="t-ja"><span class="fg-en">Sakana Namazu</span>は日本文化に対する深い理解と高性能な推論能力を兼ね備えたLLMです。Web検索やコード実行を組み合わせて様々なビジネスシーンにおける複雑なタスクをこなすことができます。今お使いのモデルから簡単に切り替えることができます。ぜひお試しください。</span>
              </p>
              <div class="fg-hero__cta">
                <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark">
                  <span class="t-en">Start Building</span>
                  <span class="t-ja">今すぐはじめる</span>
                </a>
                <a href="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-ja="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-en="https://forms.gle/HtmuMmLAGbeMS61E6" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Enterprise</span>
                  <span class="t-ja">法人プラン</span>
                </a>
              </div>
              <p class="fg-hero__notice">
                <span class="t-en">Not yet available in the EU/EEA, the UK, or Switzerland while we work toward compliance with GDPR and other region-specific regulations.</span>
                <span class="t-ja">GDPR等のEU/EEA固有規制への対応を進めており、現在はEU・EEA域内、イギリス、スイスではご利用いただけません。</span>
              </p>
            </div>
            <div class="fg-hero__media">
              <!-- アイコンモーション（ロゴの魚が泳ぐ装飾動画・白背景でページに溶け込むため枠なし）。
                   fugu 同様の環境ループ再生（音声トラックは除去済み）。差し替え時はファイル置換のみで可 -->
              <div class="nz-herowrap">
                <video class="nz-herovideo" autoplay loop muted playsinline aria-hidden="true" src="/assets/namazu-lp/namazu_icon_motion_fix_01_compressed.mp4"></video>
              </div>
            </div>
          </div>
        </section>
        <!-- §00.5 信頼帯（"also available on"。fugu §00.5 と同構造）。
             掲載は提供開始が確認できた第三者のみ: OpenRouter（モデルページ公開済み）と Merge。
             2つ目の fg-partners__logos は sakana-namazu-lp.js §4 のマーキー用クローン（表示側と同内容を保つこと） -->
        <section class="fg-partners" aria-label="Availability" data-partners-strip="">
          <div class="fg-partners__inner" data-partners-inner="">
            <span class="fg-partners__label" data-partners-label="">
              <span class="t-en">Sakana Namazu is also available on</span>
              <span class="t-ja">Sakana Namazu is also available on</span>
            </span>
            <div class="fg-partners__vp" data-partners-vp="">
              <div class="fg-partners__track" data-partners-track="">
                <div class="fg-partners__logos" data-partners-set="">
                  <a class="fg-partners__logo fg-partners__logo--or" href="https://openrouter.ai/sakana/sakana-namazu" target="_blank" rel="noreferrer" aria-label="OpenRouter">
                    <img src="/assets/namazu-lp/third-party/openrouter.png" alt="OpenRouter">
                  </a>
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://x.com/merge_api/status/2084291766375620835" target="_blank" rel="noreferrer" aria-label="Merge">
                    <img src="/assets/namazu-lp/third-party/merge.png" alt="Merge">
                  </a>
                </div>
                <div class="fg-partners__logos" data-partners-set="" aria-hidden="true" style="display:none;">
                  <a class="fg-partners__logo fg-partners__logo--or" href="https://openrouter.ai/sakana/sakana-namazu" target="_blank" rel="noreferrer" tabindex="-1" aria-label="OpenRouter">
                    <img src="/assets/namazu-lp/third-party/openrouter.png" alt="">
                  </a>
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://x.com/merge_api/status/2084291766375620835" target="_blank" rel="noreferrer" tabindex="-1" aria-label="Merge">
                    <img src="/assets/namazu-lp/third-party/merge.png" alt="">
                  </a>
                </div>
              </div>
            </div>
          </div>
        </section>
        <!-- Vercel / opencode は提供開始が確認できたら以下を有効化（両セットに同順で追加）:
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://vercel.com/ai-gateway" target="_blank" rel="noreferrer" aria-label="Vercel">
                    <img src="/assets/namazu-lp/third-party/vercel.png" alt="Vercel">
                  </a>
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://models.dev/providers/sakana/" target="_blank" rel="noreferrer" aria-label="opencode">
                    <img src="/assets/namazu-lp/third-party/opencode.png" alt="opencode">
                  </a>
        -->
        <!-- §01 とは(fugu §01 OVERVIEW 準拠。fuguはリード下にアーキ図があるが、
             Namazuは対応するビジュアルが未定のため図なしでセル直結) -->
        <section id="what-is" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <!-- レールのラベルはナビ同様に言語へ追従(fuguは常に英語)。
                   ENは「?」の孤立改行を避けるため意図的に2行 -->
              <div class="fg-aside__label">
                <span class="t-en">
                  What is
                  <br>
                  <span class="fg-en">Sakana Namazu</span>?
                </span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Namazu</span>
                  <br>
                  とは
                </span>
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">LLM with Japanese Vibes</span>
                <span class="t-ja">日本の感性で、考え、動くLLM</span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">
                  <span class="fg-en">Sakana Namazu</span>
                  is a Japanese-specialized LLM from Sakana AI. Built on the open model Kimi K2.6 and tuned with our in-house data for Japanese and Japanese business work, designed to be an LLM that is genuinely useful in Japanese business settings.
                </span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Namazu</span>
                  は、Sakana AI が提供する日本語特化のLLMです。オープンモデル Kimi K2.6 をベースに、社内独自のデータで日本語と日本の業務にあわせてチューニングし、日本のビジネスシーンで役立つLLMを目指して作られました。
                </span>
              </p>
              <div class="fg-cells fg-cells--3" aria-label="Sakana Namazu benefits">
                <article class="fg-cell">
                  <div class="fg-cell__n">01</div>
                  <h3 class="fg-cell__t">
                    <span class="t-en">Specialized for Japanese Business Contexts</span>
                    <span class="t-ja">日本語と日本の業務文脈に特化</span>
                  </h3>
                  <p class="fg-cell__b">
                    <span class="t-en">Handles business documents and email in natural Japanese, keigo and business customs included. The more your work runs on Japanese, the more <span class="fg-en">Sakana Namazu</span> shines.</span>
                    <span class="t-ja">敬語や商習慣を踏まえた自然な日本語で、業務文書やメールをそのまま処理できます。日本語でのやり取りが中心の業務ほど、<span class="fg-en">Sakana Namazu</span>は力を発揮します。</span>
                  </p>
                </article>
                <article class="fg-cell">
                  <div class="fg-cell__n">02</div>
                  <h3 class="fg-cell__t">
                    <span class="t-en">High-Performance Problem-Solving</span>
                    <span class="t-ja">高性能な問題解決能力</span>
                  </h3>
                  <p class="fg-cell__b">
                    <span class="t-en">Ask once, and the model itself searches the web, writes and runs code, and carries complex tasks through to the end.</span>
                    <span class="t-ja">一度頼めば、モデル自身がWebを調べ、コードを書いて動かしながら、複雑なタスクを最後までこなします。</span>
                  </p>
                </article>
                <article class="fg-cell">
                  <div class="fg-cell__n">03</div>
                  <h3 class="fg-cell__t">
                    <span class="t-en">Switch with a Single Settings Change</span>
                    <span class="t-ja">切り替えは、設定を変えるだけ</span>
                  </h3>
                  <p class="fg-cell__b">
                    <span class="t-en">Change one connection setting, and the code and tools you use today just work.</span>
                    <span class="t-ja">接続先の設定を1か所変えるだけで、今お使いのコードやツールがそのまま動きます。</span>
                  </p>
                </article>
              </div>
            </div>
          </div>
        </section>
        <!-- §02 Quick Start(3ステップ＋コード例。エンドポイントURLは仮) -->
        <section id="quickstart" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">
                <span class="t-en">Quick Start</span>
                <span class="t-ja">使い方</span>
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <div class="nz-qs">
              <div class="nz-qs__col">
              <h2 class="fg-h2">
                <span class="t-en">Keep Your OpenAI SDK — Start in Minutes</span>
                <span class="t-ja">いつもの OpenAI SDK のまま、すぐ使えます</span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">
                  Get an API key and update the
                  <span class="fg-mono">base_url</span>.
                  That's all. With just a few lines changed, your existing code and tools can call
                  <span class="fg-en">Sakana Namazu</span>
                  directly.
                </span>
                <span class="t-ja">
                  API キーを取得して
                  <span class="fg-mono">base_url</span>
                  を書き換えるだけ。数行の変更で、既存のコードやツールからそのまま
                  <span class="fg-en">Sakana Namazu</span>
                  を呼び出せます。
                </span>
              </p>
              <ol class="nz-qs__steps">
                <li class="nz-qs__step">
                  <span class="n">01</span>
                  <div>
                    <h3>
                      <span class="t-en">Get an API key</span>
                      <span class="t-ja">APIキーを取得</span>
                    </h3>
                    <p>
                      <span class="t-en">Create an account in the API Console and issue a key.</span>
                      <span class="t-ja">API Console でアカウントを作成し、キーを発行。</span>
                    </p>
                  </div>
                </li>
                <li class="nz-qs__step">
                  <span class="n">02</span>
                  <div>
                    <h3>
                      <span class="t-en">Change the Settings</span>
                      <span class="t-ja">設定を変える</span>
                    </h3>
                    <p>
                      <span class="t-en">Point the SDK at Namazu and set the model to sakana-namazu — the only code change.</span>
                      <span class="t-ja">接続先を差し替え、モデル名に sakana-namazu を指定。コードの変更はそれだけです。</span>
                    </p>
                  </div>
                </li>
                <li class="nz-qs__step">
                  <span class="n">03</span>
                  <div>
                    <h3>
                      <span class="t-en">Run as usual</span>
                      <span class="t-ja">そのまま実行</span>
                    </h3>
                    <p>
                      <span class="t-en">Your existing workflow just runs. Web search and code execution work through the same API, as they are.</span>
                      <span class="t-ja">既存のワークフローがそのまま動作。Web検索やコード実行も、同じAPIからそのまま使えます。</span>
                    </p>
                  </div>
                </li>
              </ol>
              <div class="nz-qs__cta nz-qs__cta--wide">
                <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark">
                  <span class="t-en">Start Building</span>
                  <span class="t-ja">今すぐはじめる</span>
                </a>
                <a href="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-ja="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-en="https://forms.gle/HtmuMmLAGbeMS61E6" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Enterprise</span>
                  <span class="t-ja">法人プラン</span>
                </a>
              </div>
              </div>
              <!-- スマホ幅のみ、CTAはコード diff の直下(--narrow)。広幅は上の --wide が出る -->
              <div class="nz-qs__coderail">
              <!-- pre内はブロックspanが改行を作るため、diff行の前後に生の改行を入れない -->
              <div class="nz-codewrap nz-qs__code">
                <!-- 言語タブ（切替は sakana-namazu-lp.js §5）。並び・タブ名は OpenAI API Reference の
                     表記（curl / Python / JavaScript）に合わせ、既定は主要客層の Python -->
                <div class="nz-codetabs" role="tablist">
                  <button type="button" class="nz-codetab" role="tab" aria-selected="false" data-active="false" data-codetab="curl">curl</button>
                  <button type="button" class="nz-codetab" role="tab" aria-selected="true" data-active="true" data-codetab="python">Python</button>
                  <button type="button" class="nz-codetab" role="tab" aria-selected="false" data-active="false" data-codetab="javascript">JavaScript</button>
                </div>
                <!-- パネルは grid で同じマスに重ね、非表示側は visibility:hidden：
                     枠の高さが常に「両タブの最大値」になり、タブ切替・言語切替で窓が上下に揺れない -->
                <div class="nz-codepanes">
                <!-- コピーはコード面の右上に浮かせる（docs と同じ置き方）。
                     表示コード＝実行可能な最終コードをそのままコピー（sakana-namazu-lp.js §5）。
                     コピー後は緑のチェックに1.6秒切替 -->
                <button type="button" class="nz-codecopy" data-codecopy aria-label="コードをコピー / Copy code" title="Copy code">
                  <svg class="ic-copy" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"></rect><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"></path></svg>
                  <svg class="ic-check" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M20 6 9 17l-5-5"></path></svg>
                </button>
                <pre class="nz-code" data-codepane="curl" data-active="false">curl https://api.sakana.ai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_SAKANA_API_KEY" \
  -d '{
    "model": "sakana-namazu",
    "messages": [{"role": "user", "content": "こんにちは！"}]
  }'</pre>
                <pre class="nz-code" data-codepane="python" data-active="true">from openai import OpenAI

client = OpenAI(<span class="chg">    base_url="https://api.sakana.ai/v1",</span><span class="chg">    api_key="YOUR_SAKANA_API_KEY",</span>)

response = client.chat.completions.create(<span class="chg">    model="sakana-namazu",</span>    messages=[{"role": "user", "content": "こんにちは！"}],
)
print(response.choices[0].message.content)</pre>
                <pre class="nz-code" data-codepane="javascript" data-active="false">import OpenAI from "openai";

const client = new OpenAI({<span class="chg">  baseURL: "https://api.sakana.ai/v1",</span><span class="chg">  apiKey: "YOUR_SAKANA_API_KEY",</span>});

const response = await client.chat.completions.create({<span class="chg">  model: "sakana-namazu",</span>  messages: [{ role: "user", content: "こんにちは！" }],
});
console.log(response.choices[0].message.content);</pre>
                </div>
              </div>
              <div class="nz-qs__cta nz-qs__cta--narrow">
                <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark">
                  <span class="t-en">Start Building</span>
                  <span class="t-ja">今すぐはじめる</span>
                </a>
                <a href="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-ja="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-en="https://forms.gle/HtmuMmLAGbeMS61E6" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Enterprise</span>
                  <span class="t-ja">法人プラン</span>
                </a>
              </div>
              </div>
              </div>
            </div>
          </div>
        </section>
        <!-- §03 モデル・料金(単価はすべて仮。ユースケース別の試算は載せない) -->
        <section id="models" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">
                <span class="t-en">Pricing</span>
                <span class="t-ja">料金</span>
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">Pricing</span>
                <span class="t-ja">ご利用料金</span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">No monthly fee — you pay only for the text and image tokens consumed by the model and your use of web search and code execution.</span>
                <span class="t-ja">月額費用はなく、お支払いはモデルが消費したテキスト・画像のトークン量と、Web検索・コード実行を使った分だけです。</span>
              </p>
              <div class="fg-plan">
                <div class="fg-plan__head">
                  <div class="fg-plan__title fg-mono">sakana-namazu</div>
                </div>
                <div class="fg-pricegrid">
                  <div class="fg-pricecell">
                    <div class="fg-pricecell__l">
                      <span class="t-en">Input</span>
                      <span class="t-ja">入力</span>
                    </div>
                    <div class="fg-pricecell__p">$0.95</div>
                    <div class="fg-pricecell__jpy" data-usd="0.95"><span class="t-en">≈ ¥152 JPY</span><span class="t-ja">約 ¥152 JPY</span></div>
                    <div class="fg-pricecell__n">/ 1M tokens</div>
                  </div>
                  <div class="fg-pricecell">
                    <div class="fg-pricecell__l">
                      <span class="t-en">Output</span>
                      <span class="t-ja">出力</span>
                    </div>
                    <div class="fg-pricecell__p">$4.00</div>
                    <div class="fg-pricecell__jpy" data-usd="4.00"><span class="t-en">≈ ¥640 JPY</span><span class="t-ja">約 ¥640 JPY</span></div>
                    <div class="fg-pricecell__n">/ 1M tokens</div>
                  </div>
                  <div class="fg-pricecell">
                    <div class="fg-pricecell__l">
                      <span class="t-en">Cached input</span>
                      <span class="t-ja">キャッシュ済み入力</span>
                    </div>
                    <div class="fg-pricecell__p">$0.15</div>
                    <div class="fg-pricecell__jpy" data-usd="0.15"><span class="t-en">≈ ¥24 JPY</span><span class="t-ja">約 ¥24 JPY</span></div>
                    <div class="fg-pricecell__n">/ 1M tokens</div>
                  </div>
                </div>
                <p class="nz-toolshead">
                  <span class="t-en">Tool usage</span>
                  <span class="t-ja">ツール利用料</span>
                </p>
                <div class="fg-pricegrid nz-pricegrid--2">
                  <div class="fg-pricecell">
                    <div class="fg-pricecell__l">
                      <span class="t-en">Web search</span>
                      <span class="t-ja">Web検索</span>
                    </div>
                    <div class="fg-pricecell__p">$7.00</div>
                    <div class="fg-pricecell__jpy" data-usd="7.00"><span class="t-en">≈ ¥1,120 JPY</span><span class="t-ja">約 ¥1,120 JPY</span></div>
                    <div class="fg-pricecell__n">
                      <span class="t-en">/ 1,000 calls, page contents included</span>
                      <span class="t-ja">/ 1,000回・ページ本文の取得込み</span>
                    </div>
                  </div>
                  <div class="fg-pricecell">
                    <div class="fg-pricecell__l">
                      <span class="t-en">Code execution</span>
                      <span class="t-ja">コード実行</span>
                    </div>
                    <div class="fg-pricecell__p">$0.12</div>
                    <div class="fg-pricecell__jpy" data-usd="0.12"><span class="t-en">≈ ¥19 JPY</span><span class="t-ja">約 ¥19 JPY</span></div>
                    <div class="fg-pricecell__n">
                      <span class="t-en">/ hour, session persisted</span>
                      <span class="t-ja">/ 時間・セッション保持</span>
                    </div>
                  </div>
                </div>
                <p class="nz-note" data-fx-note>
                  <span class="t-en">Yen amounts are reference prices converted at US$1 = ¥160. Actual billing is in US dollars.</span>
                  <span class="t-ja">円表記は 1ドル＝160円で換算した参考価格です。実際のご請求は米ドル建てとなります。</span>
                </p>
              </div>
              <div class="nz-secta">
                <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark">
                  <span class="t-en">Start Building</span>
                  <span class="t-ja">今すぐはじめる</span>
                </a>
                <!-- 料金ページ=既存 /pricing に Namazu 節が追加される構成(proj_torafugu_app)。docs TOC が h2「Sakana Namazu」をslugifyして id=sakana-namazu を付与するため #sakana-namazu -->
                <a href="https://console.sakana.ai/pricing#sakana-namazu" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Learn More</span>
                  <span class="t-ja">料金の詳細</span>
                </a>
              </div>
            </div>
          </div>
        </section>
        <!-- §04 モデル性能(技術詳細はBlogへ誘導) -->
        <section id="capability" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">
                <span class="t-en">Capabilities</span>
                <span class="t-ja">モデル性能</span>
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">Sakana Namazu Performance</span>
                <span class="t-ja">Sakana Namazuの性能</span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">
                  Keeps the capabilities of frontier open-source models, with answers tailored to the Japanese context. The strengths of
                  <span class="fg-en">Sakana Namazu</span>
                  and the technical details are covered in our blog.
                </span>
                <span class="t-ja">
                  フロンティアOSSの能力を保ちながらも、日本の文脈に合わせて回答。
                  <span class="fg-en">Sakana Namazu</span>
                  の強み、技術的な詳細は Blog で解説しています。
                </span>
              </p>
              <figure class="nz-fig">
                <div class="nz-fig__pad">
                  <img src="/assets/namazu-lp/namazu-benchmarks.png" alt="ベースモデル Kimi K2.6 とのベンチマーク比較" loading="lazy" decoding="async" style="display:block;margin:0 auto;">
                </div>
                <!-- 文言はブログのベンチマーク解説（2026-07-30版）準拠の要約 -->
                <figcaption class="nz-figcap">
                  <span class="t-en">Benchmarks vs. the Kimi K2.6 base — advanced reasoning and problem-solving (AIME26 for mathematical reasoning, MMLU-Pro for broad knowledge and reasoning, LiveCodeBench v6 for coding) retain the base model’s high performance, while understanding of Japanese and Japan-specific contexts (JFBench for instruction understanding and response quality in Japanese, our in-house Japanese translation benchmark, and FairPoliticsQA for neutral responses that avoid bias toward any single country’s values) surpasses the base across the board — most notably FairPoliticsQA, which jumps from 34.10% to 56.30%.</span>
                  <span class="t-ja">ベースモデル Kimi K2.6 との比較では、高度な推論・問題解決能力 (数学的推論の AIME26・幅広い知識と推論力の MMLU-Pro・コーディングの LiveCodeBench v6) はベースの高い性能を維持しつつ、日本語・日本固有の文脈への理解力 (日本語の指示理解・応答品質の JFBench・日本語翻訳の社内ベンチマーク Translation・特定の国の価値観に偏らない中立な回答の FairPoliticsQA) はすべてでベースを上回りました。特に FairPoliticsQA は 34.10% から 56.30% へと大幅に向上しています。</span>
                </figcaption>
              </figure>
              <!-- ブラインド比較デモ（同一設問への Model A/B/C 匿名比較）は良い設問例が
                   見つかるまで取り下げ（2026-07-27）。復活時は nz-fig + nz-placeholder で再掲 -->
              <div class="nz-secta">
                <!-- 仮リンク: ブログ記事のslug未確定（sakana-namazu-api の可能性もあり） -->
                <a href="https://sakana.ai/namazu-api" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Read our Blog</span>
                  <span class="t-ja">ブログを読む</span>
                </a>
              </div>
            </div>
          </div>
        </section>
        <!-- §04.5 利用実績（プロトタイプ／内容はダミー。事例が1件のため1社フィーチャー構成） -->
        <section id="works" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">
                <span class="t-en">Case Study</span>
                <span class="t-ja">導入事例</span>
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">Case Study</span>
                <span class="t-ja">導入事例</span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">A first look at how <span class="fg-en">Sakana Namazu</span> is being used today.</span>
                <span class="t-ja">実際に <span class="fg-en">Sakana Namazu</span> をお使いいただいている事例をご紹介します。</span>
              </p>

              <div class="nz-cs">
                <div class="nz-cs__main">
                  <div class="nz-cs__logo">
                    <img src="aillis-logo.png" alt="Aillis" width="2646" height="779" loading="lazy" decoding="async">
                    <span class="nz-cs__co">アイリス株式会社</span>
                  </div>
                  <div class="nz-cs__meta">HEALTHCARE / エビデンス検索</div>
                  <h3 class="nz-cs__t">医師のためのエビデンス検索ツールに必要な「日本の感性」</h3>
                  <p class="nz-cs__b">
                    医療AIのアイリス株式会社が提供する医師向けエビデンス検索ツール
                    <a href="https://overview.evidencefinder.jp/" target="_blank" rel="noreferrer" class="fg-link">Evidence Finder</a>
                    に、<span class="fg-en">Sakana Namazu</span> が採用されました。Evidence Finder は、医師の質問に対して PubMed などのデータベースを元に文献を検索し、引用元を明示して回答するツールです。<span class="nz-hl">正確な情報提供が求められる医療現場に対応するため、優れた日本語処理能力を持つ <span class="fg-en">Sakana Namazu</span> を活用しています</span>。加えて、引用文献が実在するかを自動検証する Verify 機能を搭載し、出力の信頼性を高めています。
                  </p>
                  <p class="nz-cs__b">
                    日本語特化の高度な言語理解力に加え、画像問題にも対応する <span class="fg-en">Sakana Namazu</span> の実力は、実績にも表れています。Evidence Finder は、2025年2月の第119回医師国家試験で正解率 <span class="nz-hl">96.8%</span>、2026年2月の第120回で <span class="nz-hl">96.4%</span> を記録しました（※1）。
                  </p>
                  <p class="nz-cs__b">
                    アイリス社が公開情報をもとに確認する限り、経済産業省 GENIAC が指定する国産基盤モデル同士の比較においても、最高得点を獲得しています（※2）。
                  </p>
                  <p class="nz-cs__fn">
                    ※1：画像問題を除いてアイリス株式会社が実施した試験下での評価です。<br>
                    ※2：2026年9月1日時点、アイリス株式会社調べ。経済産業省 GENIAC が指定する国産基盤モデルについて、各社公表資料等で確認可能な第118回以降の医師国家試験の評価結果を比較。
                  </p>
                  <div class="nz-cs__stats">
                    <div class="nz-cs__stat"><div class="nz-cs__num">96.8%</div><div class="nz-cs__cap">第119回 医師国家試験の正解率</div></div>
                    <div class="nz-cs__stat"><div class="nz-cs__num">96.4%</div><div class="nz-cs__cap">第120回 医師国家試験の正解率</div></div>
                  </div>
                </div>
                </div>

              <p class="nz-cs__note">
                <span class="t-ja">※ 他の活用事例も順次公開予定です。</span>
                <span class="t-en">More case studies coming soon.</span>
              </p>

              <div class="nz-secta">
                <a href="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-ja="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-en="https://forms.gle/HtmuMmLAGbeMS61E6" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Talk to us about your use case</span>
                  <span class="t-ja">導入について相談する</span>
                </a>
              </div>
            </div>
          </div>
        </section>

        <!-- §05 ユースケース(Case1=日本語特化×一気通貫＋低単価、Case2=エージェント能力、
             Case3=個人開発者向けの導線(Codex の UI を手作業で再現し、設定切替の story を見せる)。3件ともデモアニメ実装済み) -->
        <section id="use-cases" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">
                <span class="t-en">Use Cases</span>
                <span class="t-ja">ユースケース</span>
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">
                  What You Can Do with
                  <span class="fg-en">Sakana Namazu</span>
                </span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Namazu</span>
                  でできること
                </span>
              </h2>
              <p class="fg-lead">
                <span class="t-en"><span class="fg-en">Sakana Namazu</span> shows its strength across a wide range of business use cases and personal development projects. Here are a few examples.</span>
                <span class="t-ja">様々なビジネスユースケース、個人開発において<span class="fg-en">Sakana Namazu</span>は能力を発揮することができます。以下にいくつかのユースケースをお見せします。</span>
              </p>
              <div class="nz-ucs">
                <article class="nz-uc nz-uc--wide">
                  <div class="nz-uc__body">
                    <div class="nz-uc__tag">CASE 01</div>
                    <h3 class="nz-uc__t">
                      <span class="t-en">The Monday-Morning Market Research Report</span>
                      <span class="t-ja">毎週月曜朝の市場調査レポート</span>
                    </h3>
                  </div>
                  <div class="nz-uc__media">
                    <!-- DOM実装のデモ(動画代替案の試作)：Sakana Namazu API をバックボーンLLMにして社内エンジニアが
                         自作した想定の「週次調査エージェント」。デスクトップ風の画面に、受信トレイ(表側・左上の
                         大きい窓)とエージェント実行ログ(裏側・右下の小さい窓)の2枚が独立したウィンドウとして
                         重なる。PCと同じ作法で、後ろのウィンドウをクリックすると前面へ浮き上がって切り替わる。
                         中央の▶ボタンで再生開始(約46秒＋ツールチップ停止4回分・通常×1.6/早送り区間×0.8。再生中は ⏸／▶ で一時停止・再開、終了後は ↻ もう一度。
                         要所でデモが自動停止し、説明対象をスポットライトで残して他を暗転、ツールチップを重ねる＝オンボーディング式。再生が終わるまでデモ内は操作ロック)：
                         指示バーが開いてユーザーの指示を一度だけ登録 → 日曜夜の初回実行(一晩分のトレース)→
                         月曜朝に第1号が届き、開いて読む(自動スクロール)→ 以後の3週間を早送り
                         (裏でエージェントが毎週一晩分走り、日付の区切り線つきのトレースが高速で流れる。毎週月曜に届き続ける)。
                         再生後は support（サポートデモ）と同じくインタラクティブ：メール行クリックで各週のレポートを開ける。
                         行・レポート本文・トレースは research.js が REPORTS / TRACE データから描画
                         (t-en/t-ja ペアで言語トグルに追従)。コストは下部タスクバーの累計ティッカーのみ
                         (価格は脇役)。[data-a2-cost] の span を削除すればコスト表示だけを丸ごと外せる。
                         レポート本文とトレースの計画・検索クエリ・精読・コード行は、週次調査エージェントを
                         Namazu 生EPで4週分実走させた実出力(検索素材=2026年6〜7月の実在イベントの調査メモ・
                         企業名は分野レベル。ログの時刻のみ演出。再生成は scripts/agentic-probe/gen_research_agent.py)。
                         APIコストは同実走の個別実測値(¥9/11/10/10、累計¥40=$0.25) -->
                    <div class="nz-uc__stage">
                      <div class="nz-co-col nz-co-col--l"></div>
                      <div class="nz-uc__demoslot">
                    <div class="nz-anim2" data-uc-anim="research" role="group" aria-label="デモ：Sakana Namazu API をバックボーンLLMにして社内エンジニアが自作した想定の週次調査エージェント。日本語の指示を一度登録すると、毎週月曜の朝に市場調査レポートが受信トレイへ届き、その裏側でエージェントが計画・検索・集計・執筆を繰り返す様子を再生します。メール行や、後ろに見えているウィンドウをクリックして操作できます">
                      <div class="nz-anim2__in">
                        <button type="button" class="nz-uc__play" data-a2-play aria-label="デモを再生 / Play demo">
                          <span class="nz-uc__play-ic" aria-hidden="true">▶</span>
                          <span class="nz-uc__play-tx"><span class="t-en">Play demo</span><span class="t-ja">デモを再生</span></span>
                        </button>
                        <div class="nz-uc__spot" data-a2-spot aria-hidden="true"><!-- svg暗幕は research.js の initSpot が生成 --></div>
                        <div class="nz-anim2__topbar"><span class="nz-anim2__clock" data-a2-clock><span class="t-en">Fri 6/12 18:30</span><span class="t-ja">金 6/12 18:30</span></span><button type="button" class="nz-anim2__replay" data-a2-replay><span class="t-en">▶ Play</span><span class="t-ja">▶ 再生</span></button></div>
                        <div class="nz-anim2__desk" data-a2-desk="mail">
                          <!-- 表側：受信トレイ(左上の大きいウィンドウ) -->
                          <section class="nz-anim2__winx nz-anim2__winx--mail" data-a2-win="mail">
                            <div class="nz-anim2__titlebar">
                              <span class="nz-anim2__dots"><i></i><i></i><i></i></span>
                              <span class="wt"><span class="t-en">Inbox</span><span class="t-ja">受信トレイ</span></span>
                              <span class="wc"><b data-a2-mailcount>0</b></span>
                            </div>
                            <div class="nz-anim2__winbody">
                              <div class="nz-anim2__mailmain" data-a2-mview="list">
                                <div class="nz-anim2__list" data-a2-list><!-- 行は research.js が描画 --></div>
                                <div class="nz-anim2__reportview" data-a2-reportview><!-- レポートは research.js が描画 --></div>
                              </div>
                            </div>
                          </section>
                          <!-- 裏側：エージェント実行ログ(右下の小さいウィンドウ) -->
                          <section class="nz-anim2__winx nz-anim2__winx--be" data-a2-win="be">
                            <div class="nz-anim2__titlebar">
                              <span class="nz-anim2__dots"><i></i><i></i><i></i></span>
                              <span class="wt"><span class="t-en">weekly-research-agent — run log</span><span class="t-ja">weekly-research-agent — 実行ログ</span></span>
                              <span class="run" data-a2-runstate><span class="t-en">idle</span><span class="t-ja">待機中</span></span>
                            </div>
                            <div class="nz-anim2__winbody">
                              <!-- ユーザーが最初に一度だけ登録する指示。折りたたみ式のバー：
                                   プロローグで開いて打ち込まれ、確認を見せてから畳む。クリックで開閉可能 -->
                              <div class="nz-anim2__inst" data-a2-instwrap>
                                <button type="button" class="ih" data-a2-insttoggle>
                                  <span class="ihl"><span class="t-en">INSTRUCTION — runs weekly</span><span class="t-ja">指示 (毎週実行)</span></span>
                                  <span class="is"><span class="t-en">✓ registered</span><span class="t-ja">✓ 登録済み</span></span>
                                  <span class="chev" aria-hidden="true">▾</span>
                                </button>
                                <div class="ib">
                                  <div class="ibi">
                                    <div class="it" data-a2-inst><span class="t-en"></span><span class="t-ja"></span></div>
                                    <div class="ic" data-a2-instok><span class="t-en">✓ Registered · runs every Sun 23:00</span><span class="t-ja">✓ 登録完了 ・ 毎週 日曜 23:00 に実行</span></div>
                                  </div>
                                </div>
                              </div>
                              <div class="nz-anim2__flowmini" aria-hidden="true">
                                <span class="nd" data-a2-node="plan"><span class="t-en">Plan</span><span class="t-ja">計画</span></span>
                                <span class="ar">→</span>
                                <span class="nd" data-a2-node="search"><span class="t-en">Search</span><span class="t-ja">検索</span></span>
                                <span class="ar">→</span>
                                <span class="nd" data-a2-node="agg"><span class="t-en">Aggregate</span><span class="t-ja">集計</span></span>
                                <span class="ar">→</span>
                                <span class="nd" data-a2-node="write"><span class="t-en">Write</span><span class="t-ja">執筆</span></span>
                                
                              </div>
                              <div class="nz-anim2__trace" data-a2-trace><!-- ログ行は research.js が描画 --></div>
                              <div class="nz-anim2__befoot">Powered by <b class="br">Sakana Namazu</b></div>
                            </div>
                          </section>
                          <span class="nz-anim2__fly" data-a2-fly aria-hidden="true">✉</span>
                        </div>
                        <div class="nz-anim2__taskbar"><span class="cost" data-a2-cost><span class="t-en">API cost $<b data-a2-costen>0.00</b></span><span class="t-ja">APIコスト ¥<b data-a2-costja>0</b></span></span></div>
                      </div>
                    </div>
                        <!-- ツールチップ（オンボーディング式）。research.js がスポットライトと同時に説明対象の
                             近くへ重ねる（≤1000px はデモのみ表示でカードは出さない） -->
                        <div class="nz-co" data-a2-co="1"><span class="n">1</span><span class="task"><span class="t-en">The user asks for market research, in Japanese</span><span class="t-ja">ユーザーが日本語で市場調査を依頼</span></span><span class="cap"><span class="t-en">It reads the request precisely and drafts its own <b>research plan</b></span><span class="t-ja">指示を正確に理解し、<b>調査計画</b>を自分で立てる</span></span><span class="chip"><span class="t-en">Japanese understanding</span><span class="t-ja">日本語理解</span></span></div>
                        <div class="nz-co" data-a2-co="2"><span class="n">2</span><span class="task"><span class="t-en">Gathers information broadly</span><span class="t-ja">幅広く情報を集める</span></span><span class="cap"><span class="t-en">Iterates web_search <b>repeatedly</b> to gather what it needs</span><span class="t-ja">必要な情報を集めるため <b>web_search を繰り返す</b></span></span><span class="chip"><span class="t-en">Problem-solving</span><span class="t-ja">問題解決能力</span></span></div>
                        <div class="nz-co" data-a2-co="3"><span class="n">3</span><span class="task"><span class="t-en">Analyzes the numbers it gathered</span><span class="t-ja">集めた数字を分析</span></span><span class="cap"><span class="t-en">Writes and runs <b>code</b> itself for aggregation and charts</span><span class="t-ja"><b>コードを書いて実行</b>し、集計・作図まで自分でこなす</span></span><span class="chip"><span class="t-en">Code execution</span><span class="t-ja">コード実行</span></span></div>
                        <div class="nz-co" data-a2-co="4"><span class="n">4</span><span class="task"><span class="t-en">Writes the report</span><span class="t-ja">レポートを作成</span></span><span class="cap"><span class="t-en">Structured into facts / interpretation / implications, in <b>natural Japanese</b></span><span class="t-ja"><b>自然な日本語</b>で事実／解釈／示唆に構造化して出力</span></span><span class="chip"><span class="t-en">Japanese output</span><span class="t-ja">日本語生成</span></span></div>
                      </div>
                      <div class="nz-co-col nz-co-col--r">
                      </div>
                    </div>
                    <p class="nz-anim2__note">
                      <span class="t-en">This demo shows the UI of a research agent built on the Sakana Namazu API. The research reports are based on public information from June through July 2026, and company names in the findings are generalized to the sector level.</span>
                      <span class="t-ja">この画面は、Sakana Namazu API で構築した調査エージェントのデモUIです。調査レポートは、2026年6月から7月にかけての公開情報に基づいて作成したものです。調査結果に含まれる企業名は、分野レベルに一般化されています。</span>
                    </p>
                  </div>
                  <!-- 説明文はメディア下のキャプション位置(CASE 01 と同じ縦構成) -->
                  <p class="nz-uc__b nz-uc__b--wide">
                    <span class="t-en">From planning the research, through repeated web searches and cross-checking findings, to drafting the report — it <span class="nz-hl">runs to completion</span>, no human in the loop. You register your instructions in Japanese just once, and every Monday morning a fresh report is waiting in your inbox. Advanced research work becomes a <span class="nz-hl">weekly routine</span> that runs on its own.</span>
                    <span class="t-ja">調査計画の立案から、Web検索の繰り返し、集めた情報の突き合わせ、レポートのたたき台執筆まで、<span class="nz-hl">人手を挟まず完走</span>。日本語の指示を一度登録しておくだけで、毎週月曜の朝には最新のレポートが受信トレイに届きます。これまで担当者が半日かけていた高度な調査業務が、<span class="nz-hl">勝手にまわる毎週の定常業務</span>になります。</span>
                  </p>
                </article>
                <article class="nz-uc nz-uc--wide">
                  <div class="nz-uc__body">
                    <div class="nz-uc__tag">CASE 02</div>
                    <h3 class="nz-uc__t">
                      <span class="t-en">Automated Customer Support + Order Data Analysis</span>
                      <span class="t-ja">カスタマーサポートの自動化＋受注データ分析</span>
                    </h3>
                  </div>
                  <div class="nz-uc__media">
                    <!-- DOM実装のデモ(動画代替案の試作)：Sakana Namazu API で構築した想定のサポートコンソール風UI。
                         中央の▶ボタンで再生開始(自動再生はしない・約24秒＋ツールチップ停止3回分)。再生中は ⏸／▶ で一時停止・再開、
                         終了後は ↻ もう一度。再生が終わるまでデモ内は操作ロック、終了後はメール行・サイドバーを自由に操作できる。
                         要所でデモが自動停止し、説明対象をスポットライトで残して他を暗転、ツールチップを重ねる(オンボーディング式)。
                         メール一覧・詳細の中身は uc-anim/support.js がデータから描画(t-en/t-ja ペアで出力し言語トグルに追従)。
                         人名・社名・メール・注文データと運用数値(件数・解決率・応答時間等)は架空のシナリオ値。
                         Namazu の出力として表示する文言(処理チップ・返信文・判断・月次レポート本文)と
                         1通あたりAPIコストは生EP実走の実出力・実測。月間・本日の金額は実測単価×シナリオ件数
                         (uc-anim/support.js 冒頭コメントと scripts/agentic-probe/gen_support_agent.py を参照) -->
                    <div class="nz-uc__stage">
                      <div class="nz-co-col nz-co-col--l"></div>
                      <div class="nz-uc__demoslot">
                    <div class="nz-anim1" data-uc-anim="support" role="group" aria-label="デモ：Sakana Namazu API で構築したカスタマーサポート自動応対システム。メール一覧・自動返信の詳細・月次レポートを操作して確認できます">
                      <div class="nz-anim1__in">
                        <button type="button" class="nz-uc__play" data-a1-play aria-label="デモを再生 / Play demo">
                          <span class="nz-uc__play-ic" aria-hidden="true">▶</span>
                          <span class="nz-uc__play-tx"><span class="t-en">Play demo</span><span class="t-ja">デモを再生</span></span>
                        </button>
                        <div class="nz-uc__spot" data-a1-spot aria-hidden="true"><!-- svg暗幕は support.js の initSpot が生成 --></div>
                        <div class="nz-anim1__chrome">
                          <span class="nz-anim1__dots"><i></i><i></i><i></i></span>
                          <span class="nz-anim1__title">
                            <span class="t-en">Support Console</span>
                            <span class="t-ja">サポートコンソール</span>
                          </span>
                          <button type="button" class="nz-anim1__replay" data-a1-replay>
                            <span class="t-en">▶ Play</span>
                            <span class="t-ja">▶ 再生</span>
                          </button>
                        </div>
                        <div class="nz-anim1__app">
                          <nav class="nz-anim1__side" aria-label="デモ内ナビゲーション">
                            <button type="button" data-a1-nav="inbox" data-active="true">
                              <span class="t-en">Inbox</span>
                              <span class="t-ja">受信トレイ</span>
                              <span class="c" data-a1-navcount></span>
                            </button>
                            <button type="button" data-a1-nav="report" data-active="false">
                              <span class="t-en">Monthly report</span>
                              <span class="t-ja">月次レポート</span>
                              <span class="nd" data-a1-navdot></span>
                            </button>
                            <div class="nz-anim1__sidefoot">Powered by<br><span class="br">Sakana Namazu</span></div>
                          </nav>
                          <div class="nz-anim1__main">
                            <section class="nz-anim1__view" data-a1-view="inbox" data-active="true">
                              <header class="nz-anim1__viewhead">
                                <span class="nz-anim1__viewtitle">
                                  <span class="t-en">Inbox</span>
                                  <span class="t-ja">受信トレイ</span>
                                </span>
                                <span class="nz-anim1__count">
                                  <span class="t-en"><b data-a1-count>0</b> handled today</span>
                                  <span class="t-ja"><b data-a1-count>0</b>件 処理済み (本日)</span>
                                </span>
                              </header>
                              <div class="nz-anim1__list" data-a1-list><!-- 行は support.js が描画 --></div>
                              <footer class="nz-anim1__listfoot" data-a1-batchfoot>
                                <span class="t-en">Today: <b>120</b> auto-replied · <b>7</b> escalated · API cost <b>$0.78</b></span>
                                <span class="t-ja">本日：自動返信 <b>120</b>件 ・ 担当者へ引き継ぎ <b>7</b>件 ・ APIコスト <b>¥125</b></span>
                              </footer>
                            </section>
                            <section class="nz-anim1__view" data-a1-view="detail" data-active="false"><!-- 詳細は support.js が描画 --></section>
                            <section class="nz-anim1__view" data-a1-view="report" data-active="false">
                              <header class="nz-anim1__viewhead">
                                <span class="nz-anim1__viewtitle">
                                  <span class="t-en">Monthly Report — CS &amp; Orders (July 2026)</span>
                                  <span class="t-ja">月次レポート — CS・受注分析 (2026年7月)</span>
                                </span>
                                <span class="nz-anim1__chip">
                                  <span class="t-en">Auto-generated</span>
                                  <span class="t-ja">自動生成</span>
                                </span>
                              </header>
                              <div class="nz-anim1__doc">
                                <div class="nz-anim1__tldr">
                                  <span class="nz-anim1__tldrlabel"><span class="t-en">TL;DR</span><span class="t-ja">要約</span></span>
                                  <span class="t-en">July inquiries totaled <b>3,847 cases</b>, with a first-response time of <b>38 seconds</b> and an automated resolution rate reaching a record high of <b>94.2%</b>. API usage cost was kept to approximately <b>$0.0061</b> per case, and response time was reduced by 92% compared to manual handling.</span>
                                  <span class="t-ja">7月の問い合わせは<b>3,847件</b>となり、一次応答時間は<b>38秒</b>、自動解決率は過去最高の<b>94.2%</b>を達成した。API利用コストは1件あたり<b>約¥1.0</b>に抑えられ、人手比較で応答が92%短縮されている。</span>
                                </div>
                                <div class="nz-anim1__kpis">
                                  <div class="nz-anim1__kpi">
                                    <span class="l"><span class="t-en">Emails handled</span><span class="t-ja">処理件数</span></span>
                                    <b class="v"><span data-a1-kpi>0</span><span class="t-en"></span><span class="t-ja">件</span></b>
                                    <span class="d"><span class="t-en">+12% MoM</span><span class="t-ja">前月比 +12%</span></span>
                                  </div>
                                  <div class="nz-anim1__kpi">
                                    <span class="l"><span class="t-en">Auto-resolution rate</span><span class="t-ja">自動解決率</span></span>
                                    <b class="v">94.2%</b>
                                    <span class="d good">+1.8pt</span>
                                  </div>
                                  <div class="nz-anim1__kpi">
                                    <span class="l"><span class="t-en">Avg. first response</span><span class="t-ja">平均一次応答</span></span>
                                    <b class="v"><span class="t-en">38s</span><span class="t-ja">38秒</span></b>
                                    <span class="d good">−92%</span>
                                  </div>
                                  <div class="nz-anim1__kpi">
                                    <span class="l"><span class="t-en">Cost per inquiry</span><span class="t-ja">平均処理コスト</span></span>
                                    <b class="v"><span class="t-en">$0.006</span><span class="t-ja">¥1.0</span></b>
                                    <span class="d"><span class="t-en">$23.62 / month</span><span class="t-ja">月間 ¥3,779</span></span>
                                  </div>
                                </div>
                                <div class="nz-anim1__bars">
                                  <div class="nz-anim1__bar"><span class="l"><span class="t-en">Shipping</span><span class="t-ja">配送状況</span></span><span class="track"><span class="fill" data-a1-bar data-w="38%"></span></span><span class="v"><span class="t-en">1,462 · 38%</span><span class="t-ja">1,462・38%</span></span></div>
                                  <div class="nz-anim1__bar"><span class="l"><span class="t-en">Returns &amp; exchanges</span><span class="t-ja">返品・交換</span></span><span class="track"><span class="fill" data-a1-bar data-w="26%"></span></span><span class="v"><span class="t-en">1,000 · 26%</span><span class="t-ja">1,000・26%</span></span></div>
                                  <div class="nz-anim1__bar"><span class="l"><span class="t-en">Product specs</span><span class="t-ja">商品仕様</span></span><span class="track"><span class="fill" data-a1-bar data-w="21%"></span></span><span class="v"><span class="t-en">808 · 21%</span><span class="t-ja">808・21%</span></span></div>
                                  <div class="nz-anim1__bar"><span class="l"><span class="t-en">Other</span><span class="t-ja">その他</span></span><span class="track"><span class="fill" data-a1-bar data-w="15%"></span></span><span class="v"><span class="t-en">577 · 15%</span><span class="t-ja">577・15%</span></span></div>
                                </div>
                                <div class="nz-anim1__docbody">
                                  <div class="nz-anim1__sec" data-a1-sec>
                                    <h4><span class="t-en">1. Monthly Summary</span><span class="t-ja">1. 月次サマリー</span></h4>
                                    <ul>
                                      <li><span class="t-en">July handled <b>3,847 cases</b>, a 12% increase month-over-month, responding to rising inquiry volume.</span><span class="t-ja">7月の処理件数は3,847件で前月比12%増となり、問い合わせ増加に対応した。</span></li>
                                      <li><span class="t-en">The automated resolution rate improved by 1.8 percentage points MoM to <b>94.2%</b>, setting a new record high.</span><span class="t-ja">自動解決率は94.2%と前月比1.8ポイント改善し、過去最高を記録した。</span></li>
                                      <li><span class="t-en">API cost was approximately <b>$23.62</b> for the month, or about <b>$0.0061</b> per case, maintaining high cost performance.</span><span class="t-ja">APIコストは月間約3,779円で1件あたり約1.0円と低く、高コストパフォーマンスを維持した。</span></li>
                                    </ul>
                                  </div>
                                  <div class="nz-anim1__sec" data-a1-sec>
                                    <h4><span class="t-en">2. Inquiry Trends</span><span class="t-ja">2. 問い合わせ動向</span></h4>
                                    <ul>
                                      <li><span class="t-en">Delivery status inquiries totaled <b>1,462 cases</b>, down 12% MoM, helped by shorter shipping lead times.</span><span class="t-ja">配送状況の問い合わせは1,462件で前月比12%減となり、出荷リードタイム短縮が寄与した。</span></li>
                                      <li><span class="t-en">Returns and exchanges totaled <b>1,000 cases</b>, up 9% MoM, with size-related issues after the new product launch accounting for roughly 60%.</span><span class="t-ja">返品・交換は1,000件で前月比9%増となり、新商品発売後のサイズ起因が約6割を占めた。</span></li>
                                    </ul>
                                  </div>
                                  <div class="nz-anim1__sec" data-a1-sec>
                                    <h4><span class="t-en">3. Reconciliation with Order Data</span><span class="t-ja">3. 受注データとの突き合わせ</span></h4>
                                    <ul>
                                      <li><span class="t-en">The return rate improved by 0.4 percentage points MoM to <b>2.1%</b>, indicating an overall upward trend in quality.</span><span class="t-ja">返品率は2.1%と前月比0.4ポイント改善し、全体的な品質は向上傾向にある。</span></li>
                                      <li><span class="t-en">Damage reports totaled <b>23 cases</b>, of which <b>21 cases</b> were concentrated in dinnerware set manufacturing lot <b>L-0729</b>.</span><span class="t-ja">破損報告は23件で、うち21件が食器セットの製造ロットL-0729に集中している。</span></li>
                                    </ul>
                                  </div>
                                  <div class="nz-anim1__sec" data-a1-sec>
                                    <h4><span class="t-en">4. Improvement Proposals</span><span class="t-ja">4. 改善提案</span></h4>
                                    <ul>
                                      <li><span class="t-en">Since size-related exchanges for the new product account for 60%, improvements to size labeling and consideration of a try-on feature are needed.</span><span class="t-ja">新商品のサイズ起因交換が6割を占めるため、サイズ表記の改善や試着機能の検討が必要である。</span></li>
                                      <li><span class="t-en">Because 21 damage cases are concentrated in dinnerware set lot <b>L-0729</b>, inventory verification and a review of shipping suspension are necessary.</span><span class="t-ja">食器セットのロットL-0729で破損が21件集中しているため、在庫確認と出荷停止の検討が必要である。</span></li>
                                    </ul>
                                  </div>
                                </div>
                              </div>
                            </section>
                          </div>
                        </div>
                      </div>
                    </div>
                        <!-- ツールチップ（オンボーディング式）。support.js がスポットライトと同時に説明対象の
                             近くへ重ねる（≤1000px はデモのみ表示でカードは出さない） -->
                        <div class="nz-co" data-a1-co="1"><span class="n">1</span><span class="task"><span class="t-en">Replies to inquiry emails</span><span class="t-ja">問い合わせメールに返信</span></span><span class="cap"><span class="t-en">Auto-responds in polite, <b>natural Japanese</b></span><span class="t-ja">敬語の<b>自然な日本語</b>で自動応答</span></span><span class="chip"><span class="t-en">Japanese-first</span><span class="t-ja">日本語特化</span></span></div>
                        <div class="nz-co" data-a1-co="2"><span class="n">2</span><span class="task"><span class="t-en">Handles inquiries at volume</span><span class="t-ja">大量の問い合わせを処理</span></span><span class="cap"><span class="t-en">High throughput at a <b>low unit price</b></span><span class="t-ja">高スループットを<b>低単価</b>でさばく</span></span><span class="chip"><span class="t-en">Low cost</span><span class="t-ja">低単価</span></span></div>
                        <div class="nz-co" data-a1-co="3"><span class="n">3</span><span class="task"><span class="t-en">Builds the monthly report</span><span class="t-ja">月次レポートを作成</span></span><span class="cap"><span class="t-en"><b>Cross-checks and analyzes</b> order data, then writes it up</span><span class="t-ja">受注データと<b>突合・分析</b>して自動生成</span></span><span class="chip"><span class="t-en">Analytics</span><span class="t-ja">分析</span></span></div>
                      </div>
                      <div class="nz-co-col nz-co-col--r">
                      </div>
                    </div>
                    <p class="nz-anim1__note">
                      <span class="t-en">This demo shows the UI of a support console built on the Sakana Namazu API. The names, companies, emails, order data and operational figures such as case counts, resolution rates and response times are all fictional scenario values.</span>
                      <span class="t-ja">この画面は、Sakana Namazu API で構築したサポートコンソールのデモUIです。人名・社名・メール・注文データのほか、処理件数・解決率・応答時間などの運用数値はすべて架空のシナリオ値です。</span>
                    </p>
                  </div>
                  <!-- 説明文はメディア下のキャプション位置(fugu 定性評価と同じ縦構成) -->
                  <p class="nz-uc__b nz-uc__b--wide">
                    <span class="t-en">From reading inquiry emails and replying in <span class="nz-hl">natural, polite Japanese</span>, to cross-checking order and return data and producing monthly reports — <span class="nz-hl">end to end</span>. Each of these everyday tasks is small, but across a company they add up to a serious budget line. Sakana Namazu handles high-volume routine work at a <span class="nz-hl">low unit price</span>, keeping AI spend under control.</span>
                    <span class="t-ja">問い合わせメールの読解・<span class="nz-hl">日本語として自然な敬語での返信</span>から、受注・返品データとの突き合わせ、月次の分析レポートまで<span class="nz-hl">一気通貫</span>。こうした日常のタスクは、1件ずつは小さくても、全社で積み上がると大きな予算項目になります。Sakana Namazu は大量の日常タスクを<span class="nz-hl">低単価</span>で処理し、AI 予算の膨張を足元から抑えます。</span>
                  </p>
                </article>
                <article class="nz-uc nz-uc--wide">
                  <div class="nz-uc__body">
                    <div class="nz-uc__tag">CASE 03</div>
                    <h3 class="nz-uc__t">
                      <span class="t-en">A Fish-Swarm Light Show from a Single Topic</span>
                      <span class="t-ja">お題ひとつで、魚群のライトショーを演出</span>
                    </h3>
                  </div>
                  <div class="nz-uc__media">
                    <!-- 魚群ライトショーの実演収録（お題→演目リスト構成→Web検索で絵柄選定→tool call で魚群を編成）。
                         クリック再生の通常埋め込み（autoplay/loop なし＝他の2ケースの「▶で開始」と作法を揃える）。
                         元素材は実時間収録。背景を純白化（RGB平場除算+白点截断）した白化母版
                         docs/NA-40/demo-media/namazu-light-show.mov から 1280px/30fps/crf30 で再エンコード。
                         ポスターは 10s ペンギン演目のフレーム。
                         旧 Codex デモの実装は uc-anim/codex.{js,css}（残置・index から参照なし）に残っている -->
                    <!-- 実時間・ノーカットの収録（報幕は映像内で日英併記のため1本を両言語で共用）。
                         中央▶オーバーレイは他ケースと同じ作法：クリックで再生開始し永久に隠す。
                         controls は再生開始時に JS が付与（オーバーレイ表示中に毛ガラス越しに見えるのを防ぐ） -->
                    <div class="nz-videobox">
                      <button type="button" class="nz-uc__play" data-video-play aria-label="デモを再生 / Play demo">
                        <span class="nz-uc__play-ic" aria-hidden="true">▶</span>
                        <span class="nz-uc__play-tx"><span class="t-en">Play demo</span><span class="t-ja">デモを再生</span></span>
                      </button>
                      <video playsinline preload="metadata" poster="/assets/namazu-lp/namazu-light-show-poster.jpg">
                        <source src="/assets/namazu-lp/namazu-light-show.mp4" type="video/mp4">
                      </video>
                    </div>
                    <p class="nz-anim3__note">
                      <span class="t-en">The video is a recording of an actual performance staged by an agent built on the Sakana Namazu API.</span>
                      <span class="t-ja">映像は、Sakana Namazu API で構築したエージェントによる上演を収録したものです。</span>
                    </p>
                  </div>
                  <!-- 説明文はメディア下のキャプション位置。JA原文はmentor執筆（2026-07-29版・段落分けなしの一段組） -->
                  <p class="nz-uc__b nz-uc__b--wide">
                    <span class="t-en"><span class="fg-en">Sakana Namazu</span>’s <span class="nz-hl">image-recognition and problem-solving capabilities</span> can be put to work across a wide range of use cases. This demo shows <span class="fg-en">Sakana Namazu</span> directing a swarm of 1,000 fish to form a series of motifs along the theme “the world of the sea.” An agent built on <span class="fg-en">Sakana Namazu</span> thinks about which motifs suit the theme and <span class="nz-hl">gathers reference images from the web</span>. It then reproduces each image with the fish swarm by specifying coordinate-based positions. This demo is one example of <span class="fg-en">Sakana Namazu</span>’s ability to <span class="nz-hl">think, see, and act</span>.</span>
                    <span class="t-ja"><span class="fg-en">Sakana Namazu</span> がもつ<span class="nz-hl">画像認識能力と問題解決能力</span>は、様々なユースケースに活用できます。このデモでは、「海の世界」というテーマに沿って<span class="fg-en">Sakana Namazu</span>が1,000匹の魚群を操作し、様々なモチーフを形作る様子が示されています。<span class="fg-en">Sakana Namazu</span>をベースとしたエージェントはテーマに沿ってどのようなモチーフが良いかを考え、その<span class="nz-hl">参考画像をwebから収集</span>します。その上で、座標ベースで位置を指定することでその画像を魚群を用いて再現します。このデモは<span class="fg-en">Sakana Namazu</span>の<span class="nz-hl">「考え、見て、動かす」</span>能力を活用した一例となります。</span>
                  </p>
                </article>
              </div>
            </div>
          </div>
        </section>
        <!-- §06 FAQ(2026-07-29版・7項目。Q1はFugu LP Q1準拠。旧19項目ドラフトは docs/NA-40/faq-draft-v1-19items.html に退避) -->
        <section id="faq" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">FAQ</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <div class="fg-faqs" aria-label="Frequently asked questions">
                <div data-faq-item="" data-open="false" class="fg-faq">
                  <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-1">
                    <span class="fg-faq__n">01</span>
                    <span class="fg-faq__q">
                      <span class="t-en">How do I get started with Sakana Namazu?</span>
                      <span class="t-ja">Sakana Namazu はどうやって使い始めますか？</span>
                    </span>
                    <span class="fg-faq__icon">
                      <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14"></path><path d="M5 12h14"></path></svg>
                      <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"></path></svg>
                    </span>
                  </button>
                  <div data-faq-answer="" data-open="false" class="fg-faq__ans" id="faq-answer-1">
                    <p data-open="false">
                      <span class="t-en">Sakana Namazu is available through an OpenAI-compatible API. Point your existing client or coding harness at the Sakana Namazu endpoint with your API key and start sending requests — no SDK migration required. See the <a href="#quickstart" data-scroll-target="quickstart" class="fg-link">Quick Start</a> section above for working code.</span>
                      <span class="t-ja">Sakana Namazu は OpenAI 互換 API を通じて利用できます。既存のクライアントやコーディングハーネスを、API キーとともに Sakana Namazu のエンドポイントに向けてリクエストを送るだけです。SDK の移行は必要ありません。具体的なコードは上の <a href="#quickstart" data-scroll-target="quickstart" class="fg-link">Quick Start</a> セクションをご覧ください。</span>
                    </p>
                  </div>
                </div>
                <div data-faq-item="" data-open="false" class="fg-faq">
                  <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-2">
                    <span class="fg-faq__n">02</span>
                    <span class="fg-faq__q">
                      <span class="t-en">How does Sakana Namazu differ from Sakana Fugu?</span>
                      <span class="t-ja">Sakana Fugu とは何が違いますか？</span>
                    </span>
                    <span class="fg-faq__icon">
                      <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14"></path><path d="M5 12h14"></path></svg>
                      <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"></path></svg>
                    </span>
                  </button>
                  <div data-faq-answer="" data-open="false" class="fg-faq__ans" id="faq-answer-2">
                    <p data-open="false">
                      <span class="t-en"><a href="https://sakana.ai/fugu/" target="_blank" rel="noreferrer" class="fg-link">Sakana Fugu</a> orchestrates multiple frontier models to deliver its performance, whereas Sakana Namazu is a single in-house model — refined with our own additional training — that works on its own as an API.</span>
                      <span class="t-ja"><a href="https://sakana.ai/fugu/" target="_blank" rel="noreferrer" class="fg-link">Sakana Fugu</a> が複数のフロンティアモデルを編成 (オーケストレーション) して性能を引き出すサービスであるのに対し、Sakana Namazu は独自に追加学習した単一の自社モデルが単体で動作する API サービスです。</span>
                    </p>
                  </div>
                </div>
                <div data-faq-item="" data-open="false" class="fg-faq">
                  <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-3">
                    <span class="fg-faq__n">03</span>
                    <span class="fg-faq__q">
                      <span class="t-en">Is my data used to train the model?</span>
                      <span class="t-ja">入力したデータが学習に使われることはありますか？</span>
                    </span>
                    <span class="fg-faq__icon">
                      <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14"></path><path d="M5 12h14"></path></svg>
                      <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"></path></svg>
                    </span>
                  </button>
                  <div data-faq-answer="" data-open="false" class="fg-faq__ans" id="faq-answer-3">
                    <p data-open="false">
                      <span class="t-en">By default, your inputs may be used to train and improve our models. You can opt out at any time from the <a href="https://console.sakana.ai/settings" target="_blank" rel="noreferrer" class="fg-link">Console settings</a>. See our <a href="https://console.sakana.ai/privacy-policy" target="_blank" rel="noreferrer" class="fg-link">Privacy Policy</a> for details.</span>
                      <span class="t-ja">初期設定では、ご入力いただいたデータをモデルの学習・改良に利用する場合があります。学習への利用は、<a href="https://console.sakana.ai/settings" target="_blank" rel="noreferrer" class="fg-link">Console の設定</a>からいつでもオプトアウトできます。詳細は<a href="https://console.sakana.ai/privacy-policy" target="_blank" rel="noreferrer" class="fg-link">プライバシーポリシー</a>をご確認ください。</span>
                    </p>
                  </div>
                </div>
                <div data-faq-item="" data-open="false" class="fg-faq">
                  <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-4">
                    <span class="fg-faq__n">04</span>
                    <span class="fg-faq__q">
                      <span class="t-en">Can I pay in Japanese Yen?</span>
                      <span class="t-ja">日本円建てでの支払いはできますか？</span>
                    </span>
                    <span class="fg-faq__icon">
                      <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14"></path><path d="M5 12h14"></path></svg>
                      <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"></path></svg>
                    </span>
                  </button>
                  <div data-faq-answer="" data-open="false" class="fg-faq__ans" id="faq-answer-4">
                    <p data-open="false">
                      <span class="t-en">Not at the moment — billing is denominated in US dollars. Yen-denominated payment is available with our Enterprise plan, so please reach out via <a href="https://forms.gle/HtmuMmLAGbeMS61E6" target="_blank" rel="noreferrer" class="fg-link">Contact Us</a>.</span>
                      <span class="t-ja">現時点では対応しておらず、請求は米ドル建てのみです。法人プランでは円建てのお支払いにも対応可能ですので、<a href="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" target="_blank" rel="noreferrer" class="fg-link">Contact Us</a> よりお問い合わせください。</span>
                    </p>
                  </div>
                </div>
                <div data-faq-item="" data-open="false" class="fg-faq">
                  <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-5">
                    <span class="fg-faq__n">05</span>
                    <span class="fg-faq__q">
                      <span class="t-en">Is my data processed entirely within Japan?</span>
                      <span class="t-ja">データの処理は日本国内で完結していますか？</span>
                    </span>
                    <span class="fg-faq__icon">
                      <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14"></path><path d="M5 12h14"></path></svg>
                      <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"></path></svg>
                    </span>
                  </button>
                  <div data-faq-answer="" data-open="false" class="fg-faq__ans" id="faq-answer-5">
                    <p data-open="false">
                      <span class="t-en">At present, we cannot guarantee that data processing is completed entirely within Japan.</span>
                      <span class="t-ja">現時点では、データの処理を日本国内で完結させることは保証しておりません。</span>
                    </p>
                  </div>
                </div>
                <div data-faq-item="" data-open="false" class="fg-faq">
                  <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-6">
                    <span class="fg-faq__n">06</span>
                    <span class="fg-faq__q">
                      <span class="t-en">How does Sakana Namazu differ from the base model (Kimi K2.6)?</span>
                      <span class="t-ja">ベースモデル (Kimi K2.6) とは何が違いますか？</span>
                    </span>
                    <span class="fg-faq__icon">
                      <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14"></path><path d="M5 12h14"></path></svg>
                      <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"></path></svg>
                    </span>
                  </button>
                  <div data-faq-answer="" data-open="false" class="fg-faq__ans" id="faq-answer-6">
                    <p data-open="false">
                      <span class="t-en">We start from the open model Kimi K2.6 and apply our own additional training to strengthen Japanese performance and make responses more neutral. Details will be covered on our <a href="https://sakana.ai/namazu-api" target="_blank" rel="noreferrer" class="fg-link">technical blog</a>.</span>
                      <span class="t-ja">公開モデル Kimi K2.6 をベースに、日本語性能と回答の中立性を高めるための追加学習を独自に行っています。詳細は<a href="https://sakana.ai/namazu-api" target="_blank" rel="noreferrer" class="fg-link">技術ブログ</a>でご紹介します。</span>
                    </p>
                  </div>
                </div>
                <div data-faq-item="" data-open="false" class="fg-faq">
                  <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-7">
                    <span class="fg-faq__n">07</span>
                    <span class="fg-faq__q">
                      <span class="t-en">What is Sakana Namazu best suited for?</span>
                      <span class="t-ja">どのような用途に向いていますか？</span>
                    </span>
                    <span class="fg-faq__icon">
                      <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 5v14"></path><path d="M5 12h14"></path></svg>
                      <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 12h14"></path></svg>
                    </span>
                  </button>
                  <div data-faq-answer="" data-open="false" class="fg-faq__ans" id="faq-answer-7">
                    <p data-open="false">
                      <span class="t-en">It excels at problem-solving tasks with Japanese-language instructions. It is also worth considering if you are running frontier models at high volume today and want to bring that cost down.</span>
                      <span class="t-ja">日本語の指示を含む問題解決タスクに強みがあります。また、現在フロンティアモデルを大量に活用していて、そのコストを抑えたい場合にもご検討いただけます。</span>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
        <!-- §07 Contact(fugu §contact 準拠。問い合わせフォームURLは未定のため仮) -->
        <section id="contact" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">
                <span class="t-en">Contact Us</span>
                <span class="t-ja">お問い合わせ</span>
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 style="margin:0;font-size:clamp(32px,4.4vw,60px);font-weight:700;line-height:1.18;letter-spacing:-.01em;color:#111827;">
                <span class="t-en">
                  Ready to build with<br>
                  <span class="nz-contact-nowrap"><span style="color:var(--ac);"><span class="fg-en">Sakana Namazu</span></span>?</span>
                </span>
                <span class="t-ja">
                  <span style="color:var(--ac);">
                    <span class="fg-en">Sakana Namazu</span>
                  </span>
                  を使った開発を<span class="nz-contact-nowrap">始めてみませんか？</span>
                </span>
              </h2>
              <p class="fg-lead" style="max-width:560px;">
                <span class="t-en">Just change the settings — you can start today. For enterprise use, we can discuss requirements such as billing in Japanese Yen, invoice-based payment, region selection, and zero data retention. Get in touch.</span>
                <span class="t-ja">設定を変えるだけで、今日からはじめられます。法人でのご利用では、円建て支払い・請求書払い・リージョン指定・ゼロデータ保持 (zero data retention) など、さまざまな要件に合わせてご相談いただけます。お気軽にお問い合わせください。</span>
              </p>
              <div class="nz-contact-cta" style="display:flex;gap:12px;flex-wrap:wrap;margin-top:32px;">
                <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark">
                  <span class="t-en">Start Building</span>
                  <span class="t-ja">今すぐはじめる</span>
                </a>
                <a href="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-ja="https://docs.google.com/forms/d/e/1FAIpQLSfvDrSVigMEcLNH2F0DvgHsqp2sp-6_YOpMEx7KVhJCzi7nQA/viewform" data-href-en="https://forms.gle/HtmuMmLAGbeMS61E6" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Enterprise</span>
                  <span class="t-ja">法人プラン</span>
                </a>
              </div>
            </div>
          </div>
        </section>
      </main>
      <footer class="fg-footer">
        <div class="fg-footer__inner">
          <div class="fg-footer__logo">
            <img class="nz-logo" src="/assets/namazu-lp/logo-sakana-namazu-colored.svg" alt="Sakana Namazu">
          </div>
          <nav class="fg-footer__links" aria-label="Footer links">
            <a href="https://sakana.ai/" target="_blank" rel="noreferrer">
              <span>
                <span class="t-en">Sakana AI official site</span>
                <span class="t-ja">Sakana AI 公式サイト</span>
              </span>
              <svg width="18" height="18" viewbox="0 0 24 24" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m9 18 6-6-6-6"></path>
              </svg>
            </a>
            <a href="https://sakana.ai/company-info/" target="_blank" rel="noreferrer">
              <span>
                <span class="t-en">About us</span>
                <span class="t-ja">会社情報</span>
              </span>
              <svg width="18" height="18" viewbox="0 0 24 24" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m9 18 6-6-6-6"></path>
              </svg>
            </a>
            <a href="https://sakana.ai/careers/" target="_blank" rel="noreferrer">
              <span>
                <span class="t-en">Careers</span>
                <span class="t-ja">採用情報</span>
              </span>
              <svg width="18" height="18" viewbox="0 0 24 24" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m9 18 6-6-6-6"></path>
              </svg>
            </a>
            <a href="https://console.sakana.ai/privacy-policy" target="_blank" rel="noreferrer">
              <span>
                <span class="t-en">Privacy policy</span>
                <span class="t-ja">プライバシーポリシー</span>
              </span>
              <svg width="18" height="18" viewbox="0 0 24 24" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m9 18 6-6-6-6"></path>
              </svg>
            </a>
          </nav>
          <div class="fg-footer__copy">© 2026 Sakana AI 株式会社</div>
        </div>
      </footer>
      <div class="fg-mcta">
        <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-mcta__btn">
          <span class="t-en">Start Building</span>
          <span class="t-ja">今すぐはじめる</span>
        </a>
      </div>
    </div>
    <script src="/assets/namazu-lp/sakana-namazu-lp.js"></script>
    <script src="uc-anim/support.js"></script>
    <script src="uc-anim/research.js"></script>
  </body>
</html>
