<!DOCTYPE html>
<html lang="en" class="min-h-full scroll-smooth">
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
    <meta name="viewport" content="width=device-width">
    <link rel="icon" href="/favicon.ico">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="">
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&family=Noto+Sans+JP:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
    <meta name="description" content="Sakana Fugu is a model interface for coordinating specialized AI agents through one OpenAI-compatible API.">
    <title>Sakana Fugu — Multi-Agent System as a Model</title>


    <!-- Social cards -->
    <meta property="og:title" content="Sakana Fugu — Multi-agent System as A Model">
    <meta property="og:type" content="website">
    <meta property="og:description" content="One model to command them all">
    <meta property="og:image" content="https://sakana.ai/assets/fugu_lp/ogp.png">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta name="twitter:image" content="https://sakana.ai/assets/fugu_lp/ogp.png">
    <meta property="og:url" content="https://sakana.ai/fugu/">
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:site" content="@SakanaAILabs">

    <script>
			(function () {
				try {
					var lang = localStorage.getItem("fugu-lang");
					if (lang !== "ja" && lang !== "en") {
						var nav = (navigator.language || navigator.userLanguage || "").toLowerCase();
						lang = nav.indexOf("ja") === 0 ? "ja" : "en";
					}
					if (lang === "ja") {
						document.documentElement.classList.add("lang-ja");
						document.documentElement.setAttribute("lang", "ja");
					}
				} catch (e) {}
			})();
		
    </script>
    <link rel="stylesheet" href="/assets/css/fugu_lp.css">
    <style>
			:root{
				--ac:#e10600;
				/* Latin renders in Poppins; Japanese glyphs (which Poppins lacks) fall back to
				   Noto Sans JP — matching the design — instead of the OS Hiragino, which looks
				   more compact. Keep Noto ahead of the system fonts in every JP-bearing stack. */
				--font-en:"Poppins","Noto Sans JP",-apple-system,BlinkMacSystemFont,sans-serif;
				--font-jp:"Noto Sans JP","Hiragino Sans","Hiragino Kaku Gothic ProN","Yu Gothic",sans-serif;
				--font-mono-jb:"JetBrains Mono",ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;
			}
			body{font-family:var(--font-en);color:#111827;background:#fff;-webkit-font-smoothing:antialiased;overflow-x:clip;}
			::selection{background:rgba(225,6,0,.14);}
			.fg-mono{font-family:var(--font-mono-jb);}
			/* Keep the "Sakana Fugu" wordmark in the Latin (Poppins) stack even inside
			   Japanese copy, which otherwise renders it in Noto Sans JP. */
			.fg-en{font-family:var(--font-en);}
			/* …except in the JetBrains-Mono contexts (nav links, section eyebrow labels),
			   where the English copy itself is mono — keep the wordmark mono there too. */
			.fg-navlink .fg-en,.fg-aside__label .fg-en{font-family:var(--font-mono-jb);}
			.fg-wrap{max-width:1280px;margin:0 auto;}

			/* ---- header ---- */
			/* Header: opaque at the top of the page; becomes translucent with blur + shadow
			   once scrolled. On mobile it slides up out of view when scrolling down and
			   returns when scrolling up (driven by data-scrolled / data-hidden in JS). */
			.fg-header{position:sticky;top:0;z-index:50;background:#fff;border-bottom:1px solid #e5e7eb;transition:transform .25s ease,background .2s ease,box-shadow .2s ease;}
			.fg-header[data-scrolled="true"]{background:rgba(255,255,255,.7);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);box-shadow:0 8px 24px rgba(15,23,42,.06);}
			.fg-header[data-hidden="true"]{transform:translateY(-110%);}
			/* The full-screen mobile menu is position:fixed but lives inside this header.
			   Once scrolled, the header's backdrop-filter establishes a containing block for
			   fixed descendants, which clipped the open menu to the header's height — so its
			   body collapsed and only the bar (logo + close) showed. Drop the filter while the
			   menu is open so the overlay positions against the viewport again. */
			.fg-menu-open .fg-header[data-scrolled="true"]{backdrop-filter:none;-webkit-backdrop-filter:none;}
			.fg-header__inner{max-width:1280px;margin:0 auto;padding:0 clamp(20px,4vw,48px);min-height:clamp(64px,6vw,104px);display:flex;align-items:center;justify-content:space-between;gap:16px;}
			.fg-brand{display:flex;align-items:center;gap:14px;text-decoration:none;flex:none;}
			/* The logo PNG carries ~10px of transparent padding on each side at this
			   height; negative margins trim it in layout so the fish icon aligns with
			   the content margin and the subtitle sits snug to the wordmark (per design). */
			/* max-width:none overrides the Tailwind base img{max-width:100%}, which was
			   capping the wordmark's width in the tight mobile header and squashing it. */
			.fg-brand img{height:42px;width:auto;max-width:none;display:block;margin-left:-10px;margin-right:-10px;}
			.fg-brand__sub{display:none;font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.06em;color:#6b7280;white-space:nowrap;}
			/* Brand subtitle sits next to the logo only when the nav leaves room for it.
			   The English nav is wider than the Japanese one, so it needs a larger
			   viewport (1280px, the design width) before the subtitle fits. Japanese
			   fits from ~1160px; below that the subtitle pushed the nav wide enough to
			   clip the last item ("お問い合わせ"), so hide it under 1200px for margin. */
			@media(min-width:1200px){html.lang-ja .fg-brand__sub{display:inline;}}
			@media(min-width:1280px){html:not(.lang-ja) .fg-brand__sub{display:inline;}}
			.fg-actions{display:flex;align-items:center;gap:12px;min-width:0;}
			.fg-nav{display:flex;align-items:center;gap:6px;min-width:0;}
			.fg-nav ul{display:flex;align-items:center;gap:2px;list-style:none;margin:0;padding:0;overflow-x:auto;scrollbar-width:none;}
			.fg-nav ul::-webkit-scrollbar{display:none;}
			.fg-navlink{font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.02em;color:#374151;text-decoration:none;padding:8px 10px;white-space:nowrap;background:transparent;border:0;cursor:pointer;transition:color .15s ease;}
			.fg-navlink:hover{color:var(--ac);}
			.fg-navlink[data-active="true"]{color:var(--ac);}
			/* Language toggle pill uses JetBrains Mono (weight 500, .02em) to match the
			   design (Sakana Fugu LP - C), where the language control renders in mono —
			   not the body's Poppins-first stack. */
			.fg-pill{display:inline-flex;align-items:center;gap:8px;height:38px;padding:0 16px;background:#fff;border:1px solid #e5e7eb;border-radius:999px;cursor:pointer;font-family:var(--font-mono-jb);font-size:13px;font-weight:500;letter-spacing:.02em;color:#374151;white-space:nowrap;transition:border-color .15s ease,color .15s ease;}
			/* The global `.t-ja{font-family:var(--font-jp)}` (external fugu_lp.css) otherwise
			   renders the JA-mode pill label "Japanese" in Noto Sans JP while the EN-mode
			   "English" stays mono. Keep the pill label mono in both languages. */
			.fg-pill .t-ja{font-family:var(--font-mono-jb);}
			.fg-pill:hover{border-color:var(--ac);color:var(--ac);}
			.fg-langwrap{position:relative;}
			.fg-pill__chev{color:#6b7280;transition:transform .15s ease;}
			[data-lang-toggle][aria-expanded="true"]{border-color:var(--ac);color:var(--ac);}
			[data-lang-toggle][aria-expanded="true"] .fg-pill__chev{transform:rotate(180deg);}
			.fg-langmenu{display:none;position:absolute;right:0;top:48px;min-width:160px;background:#fff;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 12px 32px rgba(15,23,42,.12);padding:8px;z-index:80;}
			.fg-langmenu[data-open="true"]{display:flex;flex-direction:column;}
			/* Language dropdown options inherit Noto Sans JP (not the body's Poppins-first
			   stack) to match the design (Sakana Fugu LP - C), where the menu renders in
			   the root Noto Sans JP. Mirrors the footer-link treatment. */
			.fg-langopt{font-family:var(--font-jp);text-align:left;padding:10px 14px;background:transparent;border:0;border-radius:8px;cursor:pointer;font-size:15px;font-weight:600;color:#374151;}
			.fg-langopt:hover{background:#f9fafb;}
			html:not(.lang-ja) [data-lang-option="en"]{color:var(--ac);}
			html.lang-ja [data-lang-option="ja"]{color:var(--ac);}
			.fg-burger{flex:none;width:44px;height:44px;border-radius:10px;background:transparent;border:0;display:flex;align-items:center;justify-content:center;cursor:pointer;padding:0;}
			[data-desktop-nav]{display:flex;}
			[data-mobile-menu-button]{display:none;}
			/* Full-screen mobile menu (Sakana Fugu LP - C). The external stylesheet shows
			   the open menu with `display:block !important`; this equally-specific +
			   important rule (declared later) upgrades it to a flex overlay. */
			[data-mobile-menu]{display:none;}
			[data-site-header][data-nav-collapsed="true"] [data-mobile-menu][data-open="true"]{display:flex!important;}
			.fg-mobilemenu{position:fixed;inset:0;z-index:200;flex-direction:column;width:auto;max-width:none;background:#fff;border:0;box-shadow:none;padding:0;overflow-y:auto;-webkit-overflow-scrolling:touch;}
			.fg-mobilemenu li{list-style:none;}
			/* Mirror the closed header bar (.fg-header__inner mobile: 44px tall, the same
			   horizontal padding, logo at 40px with the -10px trim) so the logo and the
			   top-right button stay put when the burger toggles the menu open/closed. */
			.fg-mm__bar{position:sticky;top:0;z-index:1;flex:none;display:flex;align-items:center;justify-content:space-between;min-height:44px;padding:0 clamp(20px,4vw,48px);background:#fff;}
			.fg-mm__bar img{height:40px;width:auto;max-width:none;display:block;margin-left:-10px;}
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
			/* Hamburger menu secondary groups (per marlin LP): other products + social. */
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
			html:not(.lang-ja) .fg-mm__lang[data-lang-option="en"],html.lang-ja .fg-mm__lang[data-lang-option="ja"]{color:#111827;text-decoration:underline;text-underline-offset:4px;}
			.fg-mm__sep{color:#d1d5db;font-family:var(--font-mono-jb);}
			.fg-menu-open{overflow:hidden;}
			/* The full-screen menu lives inside the sticky header's stacking context, so the
			   root-level bottom CTA (higher in the root z-order) would paint over it. Hide it
			   while the menu is open. */
			.fg-menu-open .fg-mcta{display:none!important;}

			/* Mobile sticky bottom CTA bar */
			.fg-mcta{display:none;}
			/* The design's sticky bottom CTA inherits Noto Sans JP (not Poppins) — match it. */
			.fg-mcta__btn{display:flex;align-items:center;justify-content:center;height:50px;background:#111827;color:#fff;font-family:var(--font-jp);font-size:15px;font-weight:600;text-decoration:none;border-radius:10px;}
			.fg-mcta__btn:hover{background:#1f2937;}

			/* ---- sections / rails ---- */
			.fg-sec{padding:clamp(64px,9vw,128px) clamp(20px,4vw,48px);}
			.fg-sec--alt{background:#fafbfc;}
			.fg-grid{max-width:1280px;margin:0 auto;display:grid;grid-template-columns:200px minmax(0,1fr);gap:clamp(28px,4vw,56px);}
			.fg-aside{position:sticky;top:120px;align-self:start;display:flex;flex-direction:column;align-items:flex-start;}
			.fg-aside__label{font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.14em;color:var(--ac);}
			.fg-aside__bar{width:28px;height:2px;background:var(--ac);margin-top:10px;}
			.fg-h2{margin:0;font-size:clamp(26px,3vw,36px);font-weight:700;line-height:1.4;color:#111827;}
			.fg-lead{margin:20px 0 0;font-size:16px;line-height:1.95;color:#6b7280;max-width:680px;}
			.fg-herolead{font-size:16px;line-height:1.85;}
			.fg-link{color:var(--ac);text-decoration:underline;text-underline-offset:2px;}

			/* ---- buttons ---- */
			.fg-btn{display:inline-flex;align-items:center;height:46px;padding:0 22px;font-size:14px;font-weight:600;text-decoration:none;border:1px solid transparent;cursor:pointer;}
			.fg-btn--dark{background:#111827;color:#fff;}
			.fg-btn--dark:hover{background:#1f2937;}
			.fg-btn--ghost{background:#fff;color:#111827;border-color:#d1d5db;}
			.fg-btn--ghost:hover{border-color:#9ca3af;}

			/* ---- hero ---- */
			.fg-hero{padding:clamp(56px,8vw,104px) clamp(20px,4vw,48px) 0;}
			.fg-hero__grid{max-width:1280px;margin:0 auto;display:flex;flex-wrap:wrap;gap:clamp(32px,5vw,72px);align-items:center;padding-bottom:clamp(48px,6vw,80px);}
			.fg-hero__col{flex:1 1 340px;min-width:260px;}
			.fg-hero__media{flex:1 1 300px;min-width:260px;}
			.fg-eyebrow{font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.14em;color:var(--ac);}
			.fg-h1{margin:22px 0 0;font-family:var(--font-en);font-weight:500;font-size:clamp(42px,5.6vw,72px);line-height:1.02;letter-spacing:-.03em;color:#111827;}
			.fg-hero__cta{display:flex;gap:12px;flex-wrap:wrap;margin-top:30px;}
			.fg-hero__notice{margin:12px 0 0;max-width:520px;font-size:12.5px;line-height:1.7;color:#9ca3af;}

			/* ---- availability / partners strip ---- */
			/* The logo row auto-scales by MEASURED width, not logo count (see the
			   partners-strip controller near the end of <body>, which sets [data-mode]):
			     • fits beside the label     → static centered row      (data-mode="inline")
			     • ≥6 logos, or won't fit inline but fits on its own row → label-on-top
			                                    stacked row              (data-mode="stacked")
			     • overflows a narrow strip (≤560px) with ≤6 logos → 2-col grid  (data-mode="grid")
			     • any other overflow        → infinite marquee          (data-mode="marquee")
			   Before JS runs, the default (no [data-mode]) is the static centered row. */
			.fg-partners{--fg-logos-gap:clamp(26px,4vw,46px);border-top:1px solid #e5e7eb;border-bottom:1px solid #e5e7eb;background:#fff;}
			.fg-partners__inner{max-width:1280px;margin:0 auto;padding:clamp(24px,3.4vw,34px) clamp(20px,4vw,48px);display:flex;align-items:center;justify-content:center;gap:clamp(28px,5vw,64px);flex-wrap:nowrap;}
			.fg-partners__label{display:inline-flex;align-items:center;font-family:var(--font-mono-jb);font-size:12px;letter-spacing:.14em;color:#6b7280;white-space:nowrap;flex:none;}
			/* The label text is identical in EN/JA (日英共通). The global
			   `.t-ja{font-family:var(--font-jp)}` (external fugu_lp.css) would otherwise
			   render the JA-mode label in Noto Sans JP while EN stays mono — making the
			   same English string look different on language toggle. Keep it mono in both
			   (mirrors the .fg-pill / .fg-footer__copy treatment). */
			.fg-partners__label .t-ja{font-family:var(--font-mono-jb);}
			.fg-partners__vp{min-width:0;max-width:100%;overflow:hidden;}
			.fg-partners__track{display:flex;align-items:center;justify-content:center;}
			.fg-partners__logos{display:flex;align-items:center;gap:var(--fg-logos-gap);flex:none;}
			.fg-partners__logo{display:flex;align-items:center;opacity:1;transition:opacity .15s ease;}
			.fg-partners__logo:hover{opacity:.55;}
			.fg-partners__logo img{width:auto;max-width:none;display:block;}
			.fg-partners__logo--or img{height:22px;}
			.fg-partners__logo--vc img{height:60px;margin:-18px 0;}
			/* the clone track exists only so the marquee can loop seamlessly */
			.fg-partners__logos--clone{display:none;}

			/* stacked / grid / marquee all drop the label above the logos */
			.fg-partners[data-mode="stacked"] .fg-partners__inner,
			.fg-partners[data-mode="grid"] .fg-partners__inner,
			.fg-partners[data-mode="marquee"] .fg-partners__inner{flex-direction:column;gap:20px;}
			.fg-partners[data-mode="grid"] .fg-partners__vp,
			.fg-partners[data-mode="marquee"] .fg-partners__vp{width:100%;}
			.fg-partners[data-mode="grid"] .fg-partners__track,
			.fg-partners[data-mode="marquee"] .fg-partners__track{width:100%;}

			/* 2-col grid for ~6 logos on a narrow strip; an odd logo out gets its
			   grid-column set to "1 / -1" in JS so it spans + centers on the last row.
			   Scoped to the real set (":not(--clone)") so the hidden marquee clone
			   isn't turned into a second grid. */
			.fg-partners[data-mode="grid"] .fg-partners__logos:not(.fg-partners__logos--clone){display:grid;grid-template-columns:repeat(2,minmax(0,1fr));justify-items:center;align-items:center;gap:24px 32px;width:100%;}

			/* marquee — endless left scroll, faded edges, pauses on hover.
			   padding-right == the inter-logo gap so the set→clone seam is even; both
			   copies are identical width, so translateX(-50%) is exactly one cycle. */
			.fg-partners[data-mode="marquee"] .fg-partners__vp{-webkit-mask-image:linear-gradient(90deg,transparent,#000 7%,#000 93%,transparent);mask-image:linear-gradient(90deg,transparent,#000 7%,#000 93%,transparent);}
			.fg-partners[data-mode="marquee"] .fg-partners__track{width:max-content;justify-content:flex-start;animation:fg-pmarquee var(--fg-marquee-dur,30s) linear infinite;}
			.fg-partners[data-mode="marquee"] .fg-partners__logos{padding-right:var(--fg-logos-gap);}
			.fg-partners[data-mode="marquee"] .fg-partners__logos--clone{display:flex;}
			.fg-partners[data-mode="marquee"]:hover .fg-partners__track{animation-play-state:paused;}
			@keyframes fg-pmarquee{from{transform:translateX(0);}to{transform:translateX(-50%);}}
			@media(prefers-reduced-motion:reduce){.fg-partners__track{animation:none!important;}}

			@media(max-width:760px){
				.fg-partners{--fg-logos-gap:30px;}
				.fg-partners__inner{flex-direction:column;gap:20px;padding-top:30px;padding-bottom:30px;}
			}

			/* ---- cells (benefits / models) ---- */
			.fg-cells{display:grid;gap:1px;background:#e5e7eb;border:1px solid #e5e7eb;margin-top:clamp(32px,4vw,48px);}
			.fg-cells--3{grid-template-columns:repeat(auto-fit,minmax(220px,1fr));}
			.fg-cells--2{grid-template-columns:repeat(auto-fit,minmax(300px,1fr));}
			.fg-cells--models{max-width:860px;}
			.fg-cell{background:#fff;padding:26px 24px;}
			.fg-cell__n{font-family:var(--font-mono-jb);font-size:12px;color:var(--ac);letter-spacing:.06em;}
			.fg-cell__t{font-size:16px;font-weight:600;color:#111827;margin-top:14px;line-height:1.5;}
			.fg-cell__b{margin:10px 0 0;font-size:13.5px;line-height:1.85;color:#6b7280;}

			/* ---- figure ---- */
			.fg-figure{margin:clamp(36px,4vw,52px) 0 0;border:1px solid #e5e7eb;background:#fff;}
			.fg-figure__pad{padding:clamp(20px,3vw,40px);}
			.fg-figure img{width:100%;max-width:640px;display:block;margin:0 auto;}
			.fg-figcap{font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.06em;color:#9ca3af;border-top:1px solid #e5e7eb;padding:12px 16px;}

			/* ---- research foundations ---- */
			.fg-foundations{margin-top:clamp(32px,4vw,48px);border:1px solid #e5e7eb;}
			.fg-foundation{display:flex;flex-wrap:wrap;gap:clamp(20px,3vw,40px);align-items:center;padding:clamp(20px,3vw,32px);border-bottom:1px solid #e5e7eb;}
			.fg-foundation:last-child{border-bottom:0;}
			.fg-foundation__img{flex:1 1 320px;min-width:240px;border:1px solid #e5e7eb;background:#fff;}
			.fg-foundation__img img{width:100%;display:block;}
			.fg-foundation__body{flex:1 1 320px;min-width:240px;}
			.fg-paper{display:inline-flex;align-items:center;gap:6px;font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.06em;color:var(--ac);text-decoration:none;border-bottom:1px solid var(--ac);padding-bottom:2px;}
			.fg-foundation__t{margin:10px 0 0;font-size:18px;font-weight:600;line-height:1.5;}
			.fg-foundation__t a{color:#0f172a;text-decoration:none;}
			.fg-foundation__t a:hover{color:var(--ac);}
			.fg-foundation__b{margin:12px 0 0;font-size:13.5px;line-height:1.85;color:#6b7280;}

			/* ---- model cells ---- */
			.fg-model__head{display:flex;align-items:baseline;justify-content:space-between;gap:12px;padding-bottom:14px;border-bottom:1px solid #e5e7eb;}
			.fg-model__name{font-family:var(--font-en);font-weight:600;font-size:18px;color:#111827;}
			.fg-model__tag{font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.04em;color:var(--ac);text-align:right;}
			.fg-model__b{margin:16px 0 0;font-size:13.5px;line-height:1.85;color:#6b7280;}

			/* ---- benchmark chart + table ---- */
			/* Intro paragraph, comparison chart, and chart caption sit above the legend
			   and table (design: Sakana Fugu LP - C, §04 Benchmarks). */
			.fg-benchlead{margin:18px 0 0;font-size:16px;line-height:1.85;color:#6b7280;max-width:680px;}
			.fg-benchsub{margin:clamp(32px,4vw,48px) 0 0;font-size:22px;font-weight:600;line-height:1.5;color:#111827;}
			.fg-benchpoints{margin:12px 0 0;padding-left:22px;max-width:760px;color:#6b7280;}
			.fg-benchpoints li{font-size:14px;line-height:1.85;}
			.fg-benchpoints li+li{margin-top:8px;}
			.fg-benchpoints strong{font-weight:600;color:#111827;}
			.fg-benchchart{margin-top:24px;border:1px solid #e5e7eb;background:#fff;padding:clamp(12px,2vw,24px);}
			.fg-benchchart img{width:100%;height:auto;display:block;}
			.fg-benchchart--cyber img{width:60%;margin:0 auto;}
			.fg-benchcap{margin:16px 0 0;font-size:13px;line-height:1.8;color:#6b7280;}
			.fg-note{margin:24px 0 0;font-family:var(--font-mono-jb);font-size:12px;color:#6b7280;letter-spacing:.02em;}
			.fg-tablewrap{margin-top:24px;overflow-x:auto;border:1px solid #e5e7eb;}
			.fg-table{border-collapse:collapse;width:100%;min-width:760px;}
			.fg-table thead tr{background:#111827;}
			.fg-table th{padding:14px 18px;font-family:var(--font-mono-jb);font-size:12px;font-weight:600;letter-spacing:.03em;color:#e5e7eb;text-align:center;white-space:nowrap;}
			.fg-table td{padding:13px 14px;font-size:13.5px;text-align:center;color:#374151;font-family:var(--font-mono-jb);white-space:nowrap;}
			.fg-table tbody tr{border-top:1px solid #eef2f6;}
			.fg-table tbody tr:nth-child(odd){background:#f8fafc;}
			.fg-table tbody tr:nth-child(even){background:#fff;}
			.fg-table td.fg-th{padding:13px 18px;font-size:13px;text-align:left;color:#0f172a;font-family:var(--font-jp);font-weight:600;}
			.fg-tint{background:rgba(225,6,0,.06);}
			.fg-best{font-weight:700;color:#0f172a;}
			.fg-2nd{text-decoration:underline;text-decoration-color:#94a3b8;text-underline-offset:3px;}
			.fg-tnote{margin:16px 0 0;font-family:var(--font-mono-jb);font-size:11px;line-height:1.9;color:#9ca3af;}
			.fg-tnote p{margin:0;}

			/* ---- qualitative carousel ---- */
			.fg-exviewer{margin-top:24px;border:1px solid #e5e7eb;}
			.fg-extrack{display:flex;overflow-x:auto;overflow-y:hidden;scroll-behavior:smooth;scroll-snap-type:x mandatory;scrollbar-width:none;}
			.fg-extrack::-webkit-scrollbar{display:none;}
			.fg-exslide{min-width:100%;flex:0 0 100%;scroll-snap-align:start;}
			.fg-exframe{position:relative;background:#0f172a;aspect-ratio:16/9;}
			.fg-exframe video{width:100%;height:100%;object-fit:cover;display:block;}
			.fg-extag{padding:12px 16px;border-bottom:1px solid #e5e7eb;background:#fff;font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.04em;color:#111827;}
			.fg-excap{display:none;padding:20px;border-top:1px solid #e5e7eb;}
			.fg-excap[data-active="true"]{display:flex;gap:16px;align-items:flex-start;}
			.fg-excap p{flex:1;margin:0;font-size:13.5px;line-height:1.9;color:#6b7280;}
			.fg-exnav{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:16px 20px;border-top:1px solid #e5e7eb;}
			.fg-exbtn{width:44px;height:44px;border:1px solid #d1d5db;border-radius:6px;background:#fff;color:#111827;cursor:pointer;display:flex;align-items:center;justify-content:center;}
			.fg-exbtn:hover{border-color:#9ca3af;}
			.fg-exdots{display:flex;align-items:center;gap:8px;}
			.fg-exdot{width:9px;height:9px;border:0;border-radius:999px;cursor:pointer;padding:0;background:#cbd5e1;}
			.fg-exdot[data-active="true"]{background:var(--ac);}

			/* ---- field notes ---- */
			.fg-notes{border-top:1px solid #e5e7eb;margin-top:clamp(28px,3vw,40px);}
			.fg-note-row{display:grid;grid-template-columns:48px 1fr;gap:20px;padding:28px 0;border-bottom:1px solid #e5e7eb;}
			.fg-note-row__n{font-family:var(--font-mono-jb);font-size:13px;color:var(--ac);}
			.fg-note-row__role{font-family:var(--font-mono-jb);font-size:11px;letter-spacing:.06em;color:#9ca3af;}
			.fg-note-row__t{font-size:15px;font-weight:600;color:#111827;margin-top:8px;}
			.fg-note-row__q{margin:10px 0 0;font-size:14px;line-height:1.9;color:#374151;max-width:760px;}

			/* ---- pricing ---- */
			.fg-plan{border:1px solid #e5e7eb;background:#fff;padding:clamp(22px,3vw,32px);}
			.fg-plan + .fg-plan{margin-top:24px;}
			.fg-plan__head{display:flex;justify-content:space-between;align-items:baseline;gap:12px;padding-bottom:16px;border-bottom:1px solid #e5e7eb;}
			.fg-plan__title{font-size:18px;font-weight:700;color:#111827;margin-left:8px;}
			.fg-tag{font-family:var(--font-mono-jb);font-size:11px;padding:4px 10px;border:1px solid #111827;color:#111827;white-space:nowrap;}
			.fg-tag--ac{border-color:var(--ac);color:var(--ac);}
			.fg-plan__desc{margin:14px 0 0;font-size:13.5px;line-height:1.8;color:#6b7280;}
			.fg-subcells{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px;margin-top:18px;align-items:stretch;}
			.fg-subcells--stack{grid-template-columns:1fr;}
			.fg-subcell{padding:20px;border:1px solid #e5e7eb;background:#fafbfc;container-type:inline-size;}
			.fg-subcell__name{font-family:var(--font-en);font-size:16px;font-weight:600;color:#111827;}
			.fg-subcell__sub{font-family:var(--font-mono-jb);font-size:11px;color:#9ca3af;margin-top:4px;}
			.fg-subcell__lbl{margin-top:16px;font-size:14px;font-weight:700;color:#111827;}
			.fg-subcell__p{margin:6px 0 0;font-size:13.5px;line-height:1.85;color:#6b7280;}
			.fg-pricegrid{display:grid;grid-template-columns:repeat(3,1fr);gap:1px;background:#e5e7eb;border:1px solid #e5e7eb;margin-top:20px;}
			.fg-subcells--stack .fg-pricegrid{width:min(100%,560px);margin-top:14px;}
			@container (max-width:380px){.fg-pricegrid{grid-template-columns:1fr;}}
			.fg-pricecell{background:#fff;padding:18px 14px;}
			.fg-subcells--stack .fg-pricecell{padding:10px 10px;}
			.fg-pricecell__l{font-size:11px;color:#6b7280;}
			.fg-pricecell__p{font-family:var(--font-mono-jb);font-size:30px;font-weight:600;line-height:1.1;color:#111827;margin-top:8px;}
			.fg-subcells--stack .fg-pricecell__p{font-size:21px;margin-top:6px;}
			.fg-pricecell__p span{font-size:11px;color:#9ca3af;font-weight:400;white-space:nowrap;}
			.fg-pricecell__n{font-size:10px;line-height:1.5;color:#9ca3af;margin-top:10px;}
			.fg-tiers{display:grid;grid-template-columns:repeat(3,1fr);gap:1px;background:#e5e7eb;border:1px solid #e5e7eb;margin-top:18px;}
			.fg-tier{background:#fff;padding:20px;display:flex;flex-direction:column;}
			.fg-tier__name{font-family:var(--font-mono-jb);font-size:14px;font-weight:600;color:#111827;}
			.fg-tier__price{margin-top:10px;font-family:var(--font-mono-jb);font-size:30px;font-weight:600;letter-spacing:-.02em;color:#111827;}
			.fg-tier__price span{font-size:12px;color:#9ca3af;font-weight:400;}
			.fg-tier__sub{margin-top:10px;font-size:13px;color:#6b7280;}
			.fg-tier__desc{margin-top:6px;font-size:12px;line-height:1.7;color:#9ca3af;flex:1;}
			.fg-tier__badge{margin-top:14px;font-size:12px;font-weight:600;color:var(--ac);border-top:1px solid #e5e7eb;padding-top:10px;}

			/* ---- faq ---- */
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

			/* ---- footer ---- */
			.fg-footer{background:#fff;padding:48px clamp(20px,4vw,48px);border-top:1px solid #e5e7eb;}
			.fg-footer__inner{max-width:1280px;margin:0 auto;display:grid;grid-template-columns:1fr minmax(0,500px);grid-template-rows:1fr auto;grid-template-areas:"logo links" "copy links";column-gap:64px;row-gap:24px;min-height:280px;}
			.fg-footer__logo{grid-area:logo;align-self:end;}
			.fg-footer__logo img{height:45px;display:block;margin-left:-11px;}
			.fg-footer__links{grid-area:links;align-self:start;display:flex;flex-direction:column;}
			/* Footer links inherit Noto Sans JP (not the body's Poppins-first stack) to
			   match the design (Sakana Fugu LP - C), where the footer renders in the
			   root Noto Sans JP for both English and Japanese labels. */
			.fg-footer__links a{font-family:var(--font-jp);display:flex;align-items:center;justify-content:space-between;gap:12px;padding:20px 0;border-bottom:1px solid #e5e7eb;text-decoration:none;color:#111827;font-size:15px;transition:color .15s ease;}
			.fg-footer__links a:hover{color:var(--ac);}
			.fg-footer__links a svg{flex:none;stroke:#9ca3af;transition:stroke .15s ease;}
			.fg-footer__links a:hover svg{stroke:var(--ac);}
			.fg-footer__copy{grid-area:copy;align-self:end;font-family:var(--font-mono-jb);font-size:12px;color:#9ca3af;}
			.fg-footer__copy .t-ja{font-family:var(--font-mono-jb);}

			/* ---- responsive ---- */
			@media(max-width:760px){
				/* Compact mobile header: shorter bar, logo at 40px, and the language pill
				   removed — language is chosen at the bottom of the hamburger menu instead. */
				.fg-header__inner{min-height:44px;}
				.fg-brand img{height:40px;}
				.fg-langwrap{display:none;}
				/* minmax(0,1fr) (not bare 1fr) so wide children — the benchmark chart and
				   the min-width table — can't blow the single column past the viewport;
				   the table scrolls inside its own overflow-x wrapper instead. */
				.fg-grid{grid-template-columns:minmax(0,1fr);gap:22px;}
				.fg-table{min-width:560px;}
				/* Mobile: stack the hero into one clean full-width column (text over
				   video). Above this width the two columns stay side-by-side — small
				   enough flex-basis keeps them from wrapping, so the video always
				   fills the right half instead of leaving it blank white. */
				.fg-hero__grid{flex-direction:column;align-items:stretch;gap:clamp(24px,5vw,40px);}
				.fg-hero__col,.fg-hero__media{flex:1 1 auto;min-width:0;width:100%;}
				/* Keep both hero actions available on mobile; the sticky bar remains a
				   persistent shortcut to the primary action. */
				.fg-hero__cta{display:flex;}
				.fg-hero__cta .fg-btn{flex:1 1 0;min-width:0;justify-content:center;}
				/* Let the benchmark chart run edge-to-edge on phones (design §04). */
				.fg-benchchart{padding:0;border:none;}
				.fg-aside{position:static;top:auto;margin-bottom:16px;}
				.fg-sec{padding:48px clamp(16px,4vw,20px);}
				.fg-hero{padding:24px clamp(16px,4vw,20px) 0;}
				.fg-h1{font-size:clamp(40px,12vw,52px);}
				.fg-note-row{grid-template-columns:32px 1fr;gap:12px;}
				.fg-pricegrid,.fg-tiers{grid-template-columns:1fr;}
				.fg-faq__btn{grid-template-columns:32px 1fr 24px;gap:10px;padding:18px 2px;}
				.fg-faq__ans p{padding-left:44px;}
				.fg-mcta{display:block;position:fixed;left:0;right:0;bottom:0;z-index:55;background:rgba(255,255,255,.95);backdrop-filter:blur(10px);-webkit-backdrop-filter:blur(10px);border-top:1px solid #e5e7eb;padding:14px 16px calc(14px + env(safe-area-inset-bottom));}
				.fg-footer{padding-bottom:96px;}
				.fg-footer__inner{grid-template-columns:1fr;grid-template-rows:auto auto auto;grid-template-areas:"logo" "links" "copy";min-height:0;row-gap:0;}
				.fg-footer__links{margin-top:36px;}
				.fg-footer__copy{margin-top:36px;}
				/* SP only: shrink the hero lead paragraph (design §00: hero p 13.5px / lh1.6 on <=760).
				   All other body copy keeps its own size — only the hero is reduced. */
				.fg-herolead{font-size:13.5px;line-height:1.6;}
				/* SP: hide the hero subtitle/tagline (design §00 hides it on <=760) */
				.fg-herosub{display:none;}
			}
		
    </style>
  </head>
  <body class="min-h-full">
    <div class="min-h-screen bg-white">
      <header data-site-header="" data-nav-collapsed="true" class="fg-header">
        <div data-header-inner="" class="fg-header__inner">
          <a href="#top" data-brand="" data-scroll-target="top" class="fg-brand">
            <img src="/assets/fugu_lp/Logo_SakanaFugu.png" alt="Sakana Fugu">
            <span class="fg-brand__sub">multi-agent system as a model</span>
          </a>
          <div data-header-actions="" class="fg-actions">
            <nav data-desktop-nav="" class="fg-nav" aria-label="Primary">
              <ul>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="what-is" class="fg-navlink" aria-current="page" data-active="true">
                    <span class="t-en">
                      What is
                      <span class="fg-en">Sakana Fugu</span>
                      ?
                    </span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      とは
                    </span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="use-cases" class="fg-navlink" data-active="false">
                    <span class="t-en">How to Use</span>
                    <span class="t-ja">使い方</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="user-voices" class="fg-navlink" data-active="false">
                    <span class="t-en">Users' Voices</span>
                    <span class="t-ja">ユーザーの声</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="pricing" class="fg-navlink" data-active="false">
                    <span class="t-en">Pricing</span>
                    <span class="t-ja">料金プラン</span>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="faq" class="fg-navlink" data-active="false">
                    <span class="t-en">FAQ</span>
                    <span class="t-ja">FAQ</span>
                  </button>
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
              <button type="button" data-lang-toggle="" data-lang="en" class="fg-pill" aria-haspopup="true" aria-expanded="false" aria-controls="lang-menu" aria-label="Select language" title="Select language">
                <span class="sr-only">Select language</span>
                <span class="fg-pill__label">
                  <span class="t-en">English</span>
                  <span class="t-ja">Japanese</span>
                </span>
                <svg class="fg-pill__chev" width="14" height="14" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <path d="m6 9 6 6 6-6"></path>
                </svg>
              </button>
              <div id="lang-menu" data-lang-menu="" data-open="false" class="fg-langmenu" role="menu" aria-label="Select language">
                <button type="button" data-lang-option="ja" role="menuitemradio" aria-checked="false" class="fg-langopt">Japanese</button>
                <button type="button" data-lang-option="en" role="menuitemradio" aria-checked="true" class="fg-langopt">English</button>
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
              <img src="/assets/fugu_lp/Logo_SakanaFugu.png" alt="Sakana Fugu">
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
                        <span class="fg-en">Sakana Fugu</span>
                        ?
                      </span>
                      <span class="t-ja">
                        <span class="fg-en">Sakana Fugu</span>
                        とは
                      </span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m6 9 6 6 6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="use-cases" data-active="false">
                    <span>
                      <span class="t-en">How to Use</span>
                      <span class="t-ja">使い方</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m6 9 6 6 6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="user-voices" data-active="false">
                    <span>
                      <span class="t-en">Users' Voices</span>
                      <span class="t-ja">ユーザーの声</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m6 9 6 6 6-6"></path>
                    </svg>
                  </button>
                </li>
                <li>
                  <button type="button" data-nav-button="" data-scroll-target="pricing" data-active="false">
                    <span>
                      <span class="t-en">Pricing</span>
                      <span class="t-ja">料金プラン</span>
                    </span>
                    <svg class="fg-mm__chev" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m6 9 6 6 6-6"></path>
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
                      <path d="m6 9 6 6 6-6"></path>
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
                      <path d="m6 9 6 6 6-6"></path>
                    </svg>
                  </button>
                </li>
              </ul>
              <div class="fg-mm__divider"></div>
              <p class="fg-mm__label">Other Products</p>
              <div class="fg-mm__products">
                <a class="fg-mm__product" href="https://chat.sakana.ai/" target="_blank" rel="noreferrer">Sakana Chat</a>
                <a class="fg-mm__product" href="https://sakana.ai/marlin/" target="_blank" rel="noreferrer">Sakana Marlin</a>
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
                <button type="button" data-lang-option="ja" class="fg-mm__lang">Japanese</button>
                <span class="fg-mm__sep" aria-hidden="true">/</span>
                <button type="button" data-lang-option="en" class="fg-mm__lang">English</button>
              </div>
            </div>
          </nav>
        </div>
      </header>
      <main>
        <!-- §00 INTRODUCTION / hero -->
        <section id="top" class="fg-hero">
          <div class="fg-hero__grid">
            <div class="fg-hero__col">
              <h1 class="fg-h1">
                <span class="fg-en">Sakana Fugu</span>
              </h1>
              <div class="fg-herosub" style="margin-top:16px;font-size:clamp(16px,1.8vw,19px);font-weight:600;color:#0f172a;">
                <span class="t-en">One Model to Command Them All</span>
                <span class="t-ja">マルチエージェントを指揮する、一つのモデル</span>
              </div>
              <p class="fg-herolead" style="margin:20px 0 0;color:#6b7280;max-width:520px;">
                <span class="t-en">Frontier-level performance without single-vendor dependency. Fugu dynamically orchestrates the world's best models to tackle complex, multi-step tasks. Plug collective intelligence directly into your workflows today with a single API.</span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Fugu</span>
                  は、世界のトップモデル群を動的にオーケストレーションし、複数ステップに及ぶ複雑なタスクを自動的に解決します。高いパフォーマンスを実現するAPIを、あなたのワークフローに組み込みましょう。
                </span>
              </p>
              <div class="fg-hero__cta">
                <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark">
                  <span class="t-en">
                    Start Using
                    <span class="fg-en">Sakana Fugu</span>
                  </span>
                  <span class="t-ja">今すぐはじめる</span>
                </a>
                <a href="https://docs.google.com/forms/d/e/1FAIpQLSeCwNizCx-DZg94pNjGkxW6T_sq2TKCjv4j3O3Ds5cP0ylzhw/viewform" data-href-ja="https://docs.google.com/forms/d/e/1FAIpQLSeCwNizCx-DZg94pNjGkxW6T_sq2TKCjv4j3O3Ds5cP0ylzhw/viewform" data-href-en="https://docs.google.com/forms/d/e/1FAIpQLSc2N0uXcWW61QaDyeRKaRx5Y9ntOOI9EZ3VrhNxirunvLYrtw/viewform" target="_blank" rel="noreferrer" class="fg-btn fg-btn--ghost">
                  <span class="t-en">Contact Us</span>
                  <span class="t-ja">法人お問い合わせ</span>
                </a>
              </div>
              <p class="fg-hero__notice">
                <span class="t-en">Not yet available in the EU/EEA while we work toward compliance with GDPR and EU-specific regulations.</span>
                <span class="t-ja">GDPR等のEU/EEA固有規制への対応を進めており、現在はEU・EEA域内ではご利用いただけません。</span>
              </p>
            </div>
            <div class="fg-hero__media">
              <div style="overflow:hidden;">
                <video class="block" src="/assets/fugu_lp/videos/fugu-final-optimized-v2.mp4" autoplay="" loop="" muted="" playsinline="" preload="auto" style="display:block;width:100%;height:auto;object-fit:cover;background:#fff;" aria-label="Sakana Fugu multi-agent system animation"></video>
              </div>
            </div>
          </div>
        </section>
        <!-- §00.5 AVAILABILITY / PARTNERS -->
        <section class="fg-partners" aria-label="Availability" data-partners-strip>
          <div class="fg-partners__inner">
            <span class="fg-partners__label">
              <span class="t-en">Sakana Fugu is also available on</span>
              <span class="t-ja">Sakana Fugu is also available on</span>
            </span>
            <div class="fg-partners__vp">
              <div class="fg-partners__track">
                <div class="fg-partners__logos">
                  <a class="fg-partners__logo fg-partners__logo--or" href="https://openrouter.ai/sakana/fugu-ultra" target="_blank" rel="noreferrer" aria-label="OpenRouter">
                    <img src="/assets/fugu_lp/third-party/openrouter.png" alt="OpenRouter">
                  </a>
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://vercel.com/ai-gateway/models/fugu-ultra" target="_blank" rel="noreferrer" aria-label="Vercel">
                    <img src="/assets/fugu_lp/third-party/vercel.png" alt="Vercel">
                  </a>
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://models.dev/providers/sakana/" target="_blank" rel="noreferrer" aria-label="opencode">
                    <img src="/assets/fugu_lp/third-party/opencode.png" alt="opencode">
                  </a>
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://x.com/CreaoAI/status/2069240805496348839" target="_blank" rel="noreferrer" aria-label="Creao">
                    <img src="/assets/fugu_lp/third-party/creao.png" alt="Creao">
                  </a>
                  <a class="fg-partners__logo fg-partners__logo--vc" href="https://x.com/merge_api/status/2069433123301949763" target="_blank" rel="noreferrer" aria-label="Merge">
                    <img src="/assets/fugu_lp/third-party/merge.png" alt="Merge">
                  </a>
                </div>
              </div>
            </div>
          </div>
        </section>
        <!-- §01 OVERVIEW -->
        <section id="what-is" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">
                What is
                <span class="fg-en">Sakana Fugu</span>
                ?
              </div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">A Multi-Agent System, Delivered as One Model</span>
                <span class="t-ja">マルチエージェントを、一つのモデルAPIとして提供</span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">
                  <span class="fg-en">Sakana Fugu</span>
                  achieves superior performance by dynamically coordinating and orchestrating a diverse pool of powerful models. Instead of using domain knowledge to prescribe team organization, roles, or workflows, Fugu learns to dynamically assemble agents from a pool and coordinate them through non-obvious but highly efficient collaboration patterns.
                </span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Fugu</span>
                  は、強力で多様なモデル群を動的に組み合わせ、協調させることで高いパフォーマンスを実現します。人間が思い付かないようなモデルの編成や役割分担、処理の進め方など、効率よく学習しながら成果を発揮します。
                </span>
              </p>
              <figure class="fg-figure">
                <div class="fg-figure__pad">
                  <img src="/assets/fugu_lp/fugu_arch_image.png" width="2172" height="1606" alt="Sakana Fugu architecture overview" loading="lazy" decoding="async">
                </div>
                <div class="fg-figure__pad">
                  <img src="/assets/fugu_lp/fugu_pareto_frontier.png" width="2172" height="1606" alt="Pareto Frontier" loading="lazy" decoding="async">
                </div>
              </figure>
              <div class="fg-cells fg-cells--3" aria-label="Sakana Fugu benefits">
                <article class="fg-cell">
                  <div class="fg-cell__n">01</div>
                  <h3 class="fg-cell__t">
                    <span class="t-en">One API to Access All in an Optimized Way</span>
                    <span class="t-ja">一つのAPIで、複数モデルを最適に活用</span>
                  </h3>
                  <p class="fg-cell__b">
                    <span class="t-en">Access a coordinated pool of specialized models through one API. Fugu handles model selection and switching for each task, reducing API complexity while improving cost-performance.</span>
                    <span class="t-ja">
                      専門特化型のモデル群を、一つのAPIから利用することができます。タスクごとのモデルの選択と切り替えは
                      <span class="fg-en">Sakana Fugu</span>
                      が担うため、APIまわりの煩雑さを抑えつつ、コストパフォーマンスを高められます。
                    </span>
                  </p>
                </article>
                <article class="fg-cell">
                  <div class="fg-cell__n">02</div>
                  <h3 class="fg-cell__t">
                    <span class="t-en">Offering Superior Performance on Complex Tasks</span>
                    <span class="t-ja">複雑なタスクで優れたパフォーマンス</span>
                  </h3>
                  <p class="fg-cell__b">
                    <span class="t-en">Built for coding, reasoning, and other quality-critical workflows, Fugu coordinates expert agents to tackle complex tasks with stronger, more reliable results.</span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      は、コーディングや推論（リーズニング）など、高い品質が問われるワークフローのために設計されています。専門エージェントを連携させることで、複雑なタスクにもより確かで信頼できる答えを導きます。
                    </span>
                  </p>
                </article>
                <article class="fg-cell">
                  <div class="fg-cell__n">03</div>
                  <h3 class="fg-cell__t">
                    <span class="t-en">Providing Flexibility in Agent Selection</span>
                    <span class="t-ja">柔軟なエージェント選択</span>
                  </h3>
                  <p class="fg-cell__b">
                    <span class="t-en">Control which agents can participate in Fugu’s model pool. Opt out of specific providers or models to meet data, privacy, compliance, or organizational requirements.</span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      のモデルプールに加えるエージェントを選ぶことができます。データ、プライバシー、コンプライアンス、または組織の要件を満たすために、特定のプロバイダーやモデルを除外することが可能です。
                    </span>
                  </p>
                </article>
              </div>
            </div>
          </div>
        </section>
        <!-- §02 RESEARCH -->
        <section id="features" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">Tech Behind</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">Research-Driven Coordination for Multi-Agent Intelligence</span>
                <span class="t-ja">
                  マルチエージェントの知能を支える、
                  <br>
                  最新研究に基づく協調技術
                </span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">
                  <span class="fg-en">Sakana Fugu</span>
                  is grounded in two ICLR 2026 papers on learned model orchestration: TRINITY and the Conductor. Together, they show how systems can learn to assemble, route, and coordinate expert agents for each task instead of relying on hand-designed workflows. For a deeper look at the ideas behind the system, explore our
                  <a class="fg-link" href="https://arxiv.org/abs/2606.21228" target="_blank" rel="noreferrer">technical report</a>
                  .
                </span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Fugu</span>
                  は、モデルのオーケストレーションを学習で実現する2本のICLR 2026論文「TRINITY」と「Conductor」を基盤としています。これらの研究は、人手で設計したワークフローに頼るのではなく、タスクごとに専門エージェントをどう編成し、振り分け、連携させるかをシステム自身が学習できることを示しています。仕組みの詳細は、
                  <a class="fg-link" href="https://arxiv.org/abs/2606.21228" target="_blank" rel="noreferrer">テクニカルレポート</a>
                  をご覧ください。
                </span>
              </p>
              <div class="fg-foundations" aria-label="Technical foundations behind Sakana Fugu">
                <article class="fg-foundation">
                  <div class="fg-foundation__img">
                    <img src="/assets/fugu_lp/trinity_cover.png" width="1980" height="916" alt="Cover image for the TRINITY research paper." loading="lazy" decoding="async">
                  </div>
                  <div class="fg-foundation__body">
                    <a class="fg-paper" href="https://arxiv.org/abs/2512.04695" target="_blank" rel="noreferrer">
                      PAPER
                      <svg width="11" height="11" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                        <path d="M7 7h10v10"></path>
                        <path d="M7 17 17 7"></path>
                      </svg>
                    </a>
                    <h3 class="fg-foundation__t">
                      <a href="https://arxiv.org/abs/2512.04695" target="_blank" rel="noreferrer">
                        <span class="t-en">TRINITY: An Evolved LLM Coordinator</span>
                        <span class="t-ja">TRINITY：進化型LLMコーディネーター</span>
                      </a>
                    </h3>
                    <p class="fg-foundation__b">
                      <span class="t-en">Trinity uses a lightweight evolved coordinator to orchestrate multiple LLMs over several turns, assigning Thinker, Worker, or Verifier roles to adaptively delegate work across coding, math, reasoning, and knowledge tasks.</span>
                      <span class="t-ja">TRINITY は、軽量な進化型コーディネーターが複数のLLMを複数ターンにわたって統括する仕組み。各モデルに「Thinker（思考役）」「Worker（実行役）」「Verifier（検証役）」の役割を割り当て、コーディング・数学・推論・知識といった幅広いタスクに応じて、作業を適応的に振り分ける。</span>
                    </p>
                  </div>
                </article>
                <article class="fg-foundation">
                  <div class="fg-foundation__img">
                    <img src="/assets/fugu_lp/conductor_cover.png" width="2048" height="960" alt="Cover image for the Conductor research paper." loading="lazy" decoding="async">
                  </div>
                  <div class="fg-foundation__body">
                    <a class="fg-paper" href="https://arxiv.org/abs/2512.04388" target="_blank" rel="noreferrer">
                      PAPER
                      <svg width="11" height="11" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                        <path d="M7 7h10v10"></path>
                        <path d="M7 17 17 7"></path>
                      </svg>
                    </a>
                    <h3 class="fg-foundation__t">
                      <a href="https://arxiv.org/abs/2512.04388" target="_blank" rel="noreferrer">
                        <span class="t-en">Learning to Orchestrate Agents in Natural Language with the Conductor</span>
                        <span class="t-ja">Conductor による自然言語でのエージェント統率の学習</span>
                      </a>
                    </h3>
                    <p class="fg-foundation__b">
                      <span class="t-en">The Conductor is trained with reinforcement learning to discover natural-language coordination strategies, designing agent communication patterns and focused prompts that help diverse LLM pools outperform individual workers on challenging reasoning benchmarks.</span>
                      <span class="t-ja">Conductor は強化学習によって訓練され、自然言語ベースの協調戦略を自ら見つけ出す。エージェント間のやり取りの型や、要点を絞ったプロンプトを設計することで、多様なLLMの集まりが、難度の高い推論ベンチマークで単体のモデルを上回る力を発揮。</span>
                    </p>
                  </div>
                </article>
              </div>
            </div>
          </div>
        </section>
        <!-- §03 MODELS -->
        <section id="use-cases" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">How to Use</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">Unlock Multi-Agent Intelligence Through An API</span>
                <span class="t-ja">API を通じてマルチエージェント知能を解き放つ</span>
              </h2>
              <p class="fg-lead">
                <span class="t-en">
                  <span class="fg-en">Sakana Fugu</span>
                  comes in four models —
                  <strong style="font-weight:600;color:#111827;">Fugu</strong>,
                  <strong style="font-weight:600;color:#111827;">Fugu Ultra</strong>,
                  <strong style="font-weight:600;color:#111827;">Fugu Max</strong>, and
                  <strong style="font-weight:600;color:#111827;">Fugu Cyber</strong>
                  — all available through one OpenAI-compatible API. Pick the model that fits your workload, or switch between them without changing your integration.
                </span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Fugu</span>
                  には
                  <strong style="font-weight:600;color:#111827;">Fugu</strong>、
                  <strong style="font-weight:600;color:#111827;">Fugu Ultra</strong>、
                  <strong style="font-weight:600;color:#111827;">Fugu Max</strong>、
                  <strong style="font-weight:600;color:#111827;">Fugu Cyber</strong>
                  の 4 つのモデルがあり、いずれも OpenAI 互換 API から利用できます。ワークロードに合うモデルを選んでも、連携を変えずに切り替えてもかまいません。
                </span>
              </p>
              <div class="fg-cells fg-cells--models" aria-label="Sakana Fugu models">
                <article class="fg-cell">
                  <div class="fg-model__head">
                    <span class="fg-model__name">Fugu</span>
                    <span class="fg-model__tag">
                      <span class="t-en">Balanced performance and latency</span>
                      <span class="t-ja">性能とレイテンシのバランス</span>
                    </span>
                  </div>
                  <p class="fg-model__b">
                    <span class="t-en">Fugu balances strong performance with low latency, making it the ideal default for everyday work. Drop it into tools like Codex for coding and code review, or power responsive chatbot services — all behind a single endpoint. You can also opt specific agents out of its pool to meet data, privacy, and compliance constraints.</span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      は高い性能と低レイテンシを両立し、日々の作業に最適な標準モデルです。Codex のようなツールに組み込んでコーディングやコードレビューに使ったり、応答性の高いチャットボットを動かしたり——すべてをひとつのエンドポイントで実現します。データ・プライバシー・コンプライアンスの制約に合わせて、プールから特定のエージェントを除外することもできます。
                    </span>
                  </p>
                </article>
                <article class="fg-cell">
                  <div class="fg-model__head">
                    <span class="fg-model__name">Fugu Ultra</span>
                    <span class="fg-model__tag">
                      <span class="t-en">Optimized for performance</span>
                      <span class="t-ja">性能に最適化</span>
                    </span>
                  </div>
                  <p class="fg-model__b">
                    <span class="t-en">Fugu Ultra coordinates a deeper pool of expert agents to maximize answer quality on hard, high-stakes problems. Early users rely on it for Kaggle competitions, paper reproduction, cybersecurity analysis, and literature and patent investigations.</span>
                    <span class="t-ja">Fugu Ultra は、より広い専門エージェントのプールを連携させ、難易度が高く重要な問題で回答品質を最大化します。先行ユーザーは、Kaggle コンペティション、論文の再現、サイバーセキュリティ分析、文献・特許調査などに活用しています。</span>
                  </p>
                </article>
                <article class="fg-cell">
                  <div class="fg-model__head">
                    <span class="fg-model__name">Fugu Max</span>
                    <span class="fg-model__tag">
                      <span class="t-en">Optimized for cost-performance</span>
                      <span class="t-ja">コストパフォーマンスに最適化</span>
                    </span>
                  </div>
                  <p class="fg-model__b">
                    <span class="t-en">Fugu Max orchestrates our largest pool of models to push the cost–performance Pareto frontier. It dynamically finds efficient combinations of expert agents, aiming to improve model performance and cost efficiency together.</span>
                    <span class="t-ja">Fugu Max は、最大のモデルプールをオーケストレーションし、コストと性能のパレートフロンティアを押し広げます。専門エージェントを効率的に組み合わせることで、モデル性能とコスト効率の両方を高めることを目指します。</span>
                  </p>
                </article>
                <article class="fg-cell">
                  <div class="fg-model__head">
                    <span class="fg-model__name">Fugu Cyber</span>
                    <span class="fg-model__tag">
                      <span class="t-en">Specialized for cybersecurity</span>
                      <span class="t-ja">サイバーセキュリティに特化</span>
                    </span>
                  </div>
                  <p class="fg-model__b">
                    <span class="t-en">Fugu Cyber is specialized for cybersecurity reasoning tasks. It brings orchestration to security analysis, vulnerability research, threat investigation, and other cyber workflows where depth and precision matter.</span>
                    <span class="t-ja">Fugu Cyber はサイバーセキュリティタスクに特化したモデルです。セキュリティ分析、脆弱性調査、脅威検知など、深さと精度が重要なサイバーセキュリティワークフローにオーケストレーションを提供します。</span>
                  </p>
                </article>
              </div>
            </div>
          </div>
        </section>
        <!-- §04 BENCHMARKS -->
        <section id="quantitative-results" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">Quantitative Results</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">Quantitative Results</span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Fugu</span>
                  の性能：定量評価
                </span>
              </h2>
              <h3 class="fg-benchsub">Fugu Ultra (v2.0)</h3>
              <p class="fg-benchlead">
                <span class="t-en">For complex multi-step reasoning, autonomous research, and full-stack software development, Fugu Ultra sets our new benchmark for raw output quality. In summary, Fugu Ultra</span>
                <span class="t-ja">複雑なマルチステップ推論、自律型リサーチ、フルスタックソフトウェア開発において、Fugu Ultra は出力品質そのものに関する新たな基準を打ち立てます。Fugu Ultraの成果は以下の通りです。</span>
              </p>
              <ul class="fg-benchpoints">
                <li>
                  <span class="t-en"><strong>Performance:</strong> Achieves the best or joint-best score on five of eight benchmarks: GDP.pdf, Chartography, DeepSWE, Toolathon, and SWEFish, our internal benchmark reflecting Sakana AI’s own coding challenges and use-cases.</span>
                  <span class="t-ja"><strong>性能：</strong>8 つのベンチマークのうち、GDP.pdf、Chartography、DeepSWE、Toolathon、そして Sakana AI 独自のコーディング課題やユースケースを反映した社内ベンチマークである SWEFish の 5 つで、単独最高または同率最高のスコアを達成しています。</span>
                </li>
                <li>
                  <span class="t-en"><strong>Consistency:</strong> Places in the top 2 on seven of eight benchmarks, demonstrating strong performance across a broad range of agentic tasks.</span>
                  <span class="t-ja"><strong>一貫性：</strong>8 つのベンチマークのうち 7 つでトップ 2 に入り、幅広いエージェント型タスクで高い性能を発揮しています。</span>
                </li>
                <li>
                  <span class="t-en"><strong>Frontier:</strong> With a focus different from Fugu Max, Fugu Ultra pushes the Pareto frontier in the performance direction, providing a higher-capability option for workloads where quality is the priority.</span>
                  <span class="t-ja"><strong>フロンティア：</strong>Fugu Max とは異なる方向性で、Fugu Ultra はパレートフロンティアを性能方向へ押し広げ、品質を最優先するワークロードにより高い能力を提供します。</span>
                </li>
              </ul>
              <p class="fg-benchlead">
                <span class="t-en">Fugu Ultra does not rely on individual proprietary frontier models to deliver frontier output. By orchestrating a swappable pool of open and specialized models, it outperforms closed ecosystems while protecting users from vendor lock-in, API revocations, geopolitical turbulence, and sudden service cutoffs.</span>
                <span class="t-ja">Fugu Ultra は、フロンティアレベルの出力を実現するために、個別のプロプライエタリなフロンティアモデルへ依存していません。入れ替え可能なオープンモデルと専門モデルのプールをオーケストレーションすることで、クローズドなエコシステムを上回る性能を発揮しながら、ベンダーロックイン、API の提供停止、地政学的な不確実性、突然のサービス遮断からユーザーを保護します。</span>
              </p>
              <div class="fg-benchchart">
                <img src="/assets/fugu_lp/fugu_ultra_v2_benchmarks_2x4.png" alt="Fugu Ultra v2 benchmark results" loading="lazy" decoding="async">
              </div>
              <p class="fg-benchcap">
                <span class="t-en">Our flagship Fugu Ultra model continues to deliver a performance that is better or on-par with frontier models. Note: Fugu Ultra’s training cutoff date is 20260828, Fable 5, Fable 5.1 and GPT-6-Astra are <strong style="font-weight:700;color:#111827;">NOT</strong> in Fugu-Ultra’s model pool.</span>
                <span class="t-ja">フラッグシップモデルである Fugu Ultra は、引き続きフロンティアモデルを上回る、または同等の性能を発揮しています。注：Fugu Ultra の学習データのカットオフ日は 20260828 であり、Fable 5、Fable 5.1、GPT-6-Astra は Fugu Ultra のモデルプールに含まれていません。</span>
              </p>
              <h3 class="fg-benchsub">Fugu Max (v1.0)</h3>
              <p class="fg-benchlead">
                <span class="t-en">Fugu Max expands the pool of models Sakana Fugu can orchestrate, integrating an unprecedented number of open-weights and specialized models, including NVIDIA's Nemotron family through our collaboration with NVIDIA. It sits at a point on the Pareto frontier that single-model providers cannot reach. Concretely,</span>
                <span class="t-ja">Fugu Max は、Sakana Fugu がオーケストレーションできるモデルプールを拡大し、NVIDIA との協業を通じて利用する NVIDIA の Nemotron ファミリーを含む、これまでにない数のオープンウェイトモデルと専門モデルを統合しています。単一モデルのプロバイダーでは到達できないパレートフロンティア上の位置を実現しています。具体的には、</span>
              </p>
              <ul class="fg-benchpoints">
                <li>
                  <span class="t-en"><strong>Performance:</strong> Fugu Max achieves best overall score on six benchmarks, including Terminal Bench 2.1, GPQAD, AA-LCR, GDP.pdf, AutomationBench, and SWEFish, our internal benchmark reflecting Sakana AI’s own coding challenges and use-cases.</span>
                  <span class="t-ja"><strong>性能：</strong>Fugu Max は、Terminal Bench 2.1、GPQAD、AA-LCR、GDP.pdf、AutomationBench、そして Sakana AI 独自のコーディング課題やユースケースを反映した社内ベンチマークである SWEFish を含む 6 つのベンチマークで、総合最高スコアを達成しています。</span>
                </li>
                <li>
                  <span class="t-en"><strong>Cost:</strong> At $2 per million input tokens and $6 per million output tokens, Fugu Max’s output pricing is 40-60% lower than Sonnet 5, GPT 5.6 Terra, and Kimi K3.</span>
                  <span class="t-ja"><strong>コスト：</strong>Fugu Max は入力 100 万トークンあたり $2、出力 100 万トークンあたり $6 で、出力料金は Sonnet 5、GPT 5.6 Terra、Kimi K3 より 40〜60% 低く設定されています。</span>
                </li>
                <li>
                  <span class="t-en"><strong>Efficiency:</strong> Fugu Max expands the cost-performance Pareto frontier on seven out of ten benchmarks, delivering performance beyond the existing baseline efficiency envelope.</span>
                  <span class="t-ja"><strong>効率性：</strong>Fugu Max は 10 のベンチマークのうち 7 つでコストパフォーマンスのパレートフロンティアを押し広げ、既存のベースラインが形成する効率性の領域を超える性能を実現しています。</span>
                </li>
              </ul>
              <div class="fg-benchchart">
                <img src="/assets/fugu_lp/fugu_max_v02_output_price_scatter_5x2.png" alt="Fugu Max cost-performance Pareto frontier" loading="lazy" decoding="async">
              </div>
              <p class="fg-benchcap">
                <span class="t-en">Among frontier models in a similar price range (input prices per 1M tokens for each model are shown in the first subplot), Fugu Max expands the pareto frontier formed by single models and places itself in a cost-performance efficient position across multiple benchmarks.</span>
                <span class="t-ja">同価格帯のフロンティアモデル（各モデルの入力 100 万トークンあたりの料金は最初のサブプロットに表示）の中で、Fugu Max は単一モデル群が形成するパレートフロンティアを押し広げ、複数のベンチマークにわたってコスト-パフォーマンスの優位性を実現しています。</span>
              </p>
              <h3 class="fg-benchsub">Fugu Cyber (v1.0)</h3>
              <p class="fg-benchlead">
                <span class="t-en">For cybersecurity workflows that demand deep technical reasoning and careful investigation, Fugu Cyber applies multi-agent orchestration to security analysis, vulnerability research, and threat investigation. By coordinating specialized agents, it can sustain complex investigative workflows where precision and depth matter.</span>
                <span class="t-ja">高度な技術的推論と慎重な調査が求められるサイバーセキュリティ領域において、Fugu Cyber はマルチエージェント・オーケストレーションをセキュリティ分析、脆弱性調査、脅威調査に活用します。専門エージェントを連携させることで、精度と深さが重要となる複雑な調査ワークフローを継続的に遂行します。</span>
              </p>
              <div class="fg-benchchart fg-benchchart--cyber">
                <img src="/assets/fugu_lp/cyber_benchmark-chart.png" alt="Cybersecurity benchmark comparison chart" loading="lazy" decoding="async">
              </div>
              <p class="fg-benchcap">
                <span class="t-en">Fugu-Cyber achieves state-of-the-art performance on the industry's most challenging security benchmarks, reaching a success rate of 86.9% on CyberGym and 72.1% on CTI-REALM, comparable to leading cybersecurity-focused frontier models such as GPT-5.5-Cyber and Mythos-Preview.</span>
                <span class="t-ja">Fugu-Cyber は、業界で最も難しいセキュリティベンチマークにおいて最先端の性能を達成し、CyberGym で 86.9%、CTI-REALM で 72.1% の成功率に到達しました。これは GPT-5.5-Cyber や Mythos-Preview など、サイバーセキュリティに特化した主要なフロンティアモデルに匹敵する結果です。</span>
              </p>
            </div>
          </div>
        </section>
        <!-- §05 QUALITATIVE -->
        <section id="qualitative-results" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">Qualitative Results</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">Qualitative Results</span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Fugu</span>
                  の性能：定性的な例
                </span>
              </h2>
              <p class="fg-lead" style="max-width:760px;">
                <span class="t-en">
                  These examples compare
                  <span class="fg-en">Sakana Fugu</span>
                  models with three frontier baselines —
                  <strong style="font-weight:600;color:#2563eb;">Gemini 3.1 Pro (high)</strong>
                  ,
                  <strong style="font-weight:600;color:#ea580c;">Opus 4.8 (max)</strong>
                  , and
                  <strong style="font-weight:600;color:#dc2626;">GPT 5.5 (xhigh)</strong>
                  . To keep the focus on behavior rather than brand-by-brand attribution, the baselines are anonymized as
                  <strong style="font-weight:600;color:#374151;">Model A</strong>
                  ,
                  <strong style="font-weight:600;color:#374151;">Model B</strong>
                  , and
                  <strong style="font-weight:600;color:#374151;">Model C</strong>
                  in each description.
                  <strong style="font-weight:600;color:#374151;">The mapping is intentionally not fixed across examples.</strong>
                </span>
                <span class="t-ja">
                  以下の例では、
                  <span class="fg-en">Sakana Fugu</span>
                  を、
                  <strong style="font-weight:600;color:#2563eb;">Gemini 3.1 Pro（high）</strong>
                  、
                  <strong style="font-weight:600;color:#ea580c;">Opus 4.8 （max）</strong>
                  、
                  <strong style="font-weight:600;color:#dc2626;">GPT 5.5（xhigh）</strong>
                  の3つのフロンティアモデルと比較しています。個別モデルではなく挙動の違いに注目できるよう、ベースラインを
                  <strong style="font-weight:600;color:#374151;">Model A</strong>
                  、
                  <strong style="font-weight:600;color:#374151;">Model B</strong>
                  、
                  <strong style="font-weight:600;color:#374151;">Model C</strong>
                  として匿名化しています。
                  <strong style="font-weight:600;color:#374151;">なお、どのモデルがA〜Cかは例ごとに変えています。</strong>
                </span>
              </p>
              <div class="fg-exviewer" aria-label="Sakana Fugu qualitative examples carousel">
                <div data-examples-track="" class="fg-extrack">
                  <article class="fg-exslide">
                    <div class="fg-extag">
                      <span class="t-en">AutoResearch · BPB</span>
                      <span class="t-ja">AutoResearch · BPB</span>
                    </div>
                    <div class="fg-exframe" data-example-index="0">
                      <video data-example-video="" data-index="0" src="/assets/fugu_lp/videos/demo0.mp4" controls="" loop="" muted="" playsinline="" preload="metadata" aria-label="Sakana Fugu qualitative example 1"></video>
                      <div data-video-warning="" class="hidden" style="position:absolute;inset:0;flex-direction:column;align-items:center;justify-content:center;gap:6px;padding:0 16px;text-align:center;">
                        <span style="font-size:24px;" aria-hidden="true">⚠️</span>
                        <span style="font-size:14px;font-weight:600;color:rgba(255,255,255,.8);">
                          <span class="t-en">Video unavailable</span>
                          <span class="t-ja">動画を表示できません</span>
                        </span>
                        <span style="font-size:12px;color:rgba(255,255,255,.5);">demo0.mp4</span>
                      </div>
                    </div>
                  </article>
                  <article class="fg-exslide">
                    <div class="fg-extag">
                      <span class="t-en">Kana letters · Reading order</span>
                      <span class="t-ja">かな文書 · 読み順推定</span>
                    </div>
                    <div class="fg-exframe" data-example-index="1">
                      <video data-example-video="" data-index="1" src="/assets/fugu_lp/videos/demo1.mp4" controls="" loop="" muted="" playsinline="" preload="metadata" aria-label="Sakana Fugu qualitative example 2"></video>
                      <div data-video-warning="" class="hidden" style="position:absolute;inset:0;flex-direction:column;align-items:center;justify-content:center;gap:6px;padding:0 16px;text-align:center;">
                        <span style="font-size:24px;" aria-hidden="true">⚠️</span>
                        <span style="font-size:14px;font-weight:600;color:rgba(255,255,255,.8);">
                          <span class="t-en">Video unavailable</span>
                          <span class="t-ja">動画を表示できません</span>
                        </span>
                        <span style="font-size:12px;color:rgba(255,255,255,.5);">demo1.mp4</span>
                      </div>
                    </div>
                  </article>
                  <article class="fg-exslide">
                    <div class="fg-extag">
                      <span class="t-en">Rubik's Cube · Solver</span>
                      <span class="t-ja">Rubik’s Cube · ソルバー生成</span>
                    </div>
                    <div class="fg-exframe" data-example-index="2">
                      <video data-example-video="" data-index="2" src="/assets/fugu_lp/videos/demo2.mp4" controls="" loop="" muted="" playsinline="" preload="metadata" aria-label="Sakana Fugu qualitative example 3"></video>
                      <div data-video-warning="" class="hidden" style="position:absolute;inset:0;flex-direction:column;align-items:center;justify-content:center;gap:6px;padding:0 16px;text-align:center;">
                        <span style="font-size:24px;" aria-hidden="true">⚠️</span>
                        <span style="font-size:14px;font-weight:600;color:rgba(255,255,255,.8);">
                          <span class="t-en">Video unavailable</span>
                          <span class="t-ja">動画を表示できません</span>
                        </span>
                        <span style="font-size:12px;color:rgba(255,255,255,.5);">demo2.mp4</span>
                      </div>
                    </div>
                  </article>
                  <article class="fg-exslide">
                    <div class="fg-extag">
                      <span class="t-en">CAD · Mechanical iris</span>
                      <span class="t-ja">CAD · メカニカルアイリス</span>
                    </div>
                    <div class="fg-exframe" data-example-index="3">
                      <video data-example-video="" data-index="3" src="/assets/fugu_lp/videos/demo3.mp4" controls="" loop="" muted="" playsinline="" preload="metadata" aria-label="Sakana Fugu qualitative example 4"></video>
                      <div data-video-warning="" class="hidden" style="position:absolute;inset:0;flex-direction:column;align-items:center;justify-content:center;gap:6px;padding:0 16px;text-align:center;">
                        <span style="font-size:24px;" aria-hidden="true">⚠️</span>
                        <span style="font-size:14px;font-weight:600;color:rgba(255,255,255,.8);">
                          <span class="t-en">Video unavailable</span>
                          <span class="t-ja">動画を表示できません</span>
                        </span>
                        <span style="font-size:12px;color:rgba(255,255,255,.5);">demo3.mp4</span>
                      </div>
                    </div>
                  </article>
                  <article class="fg-exslide">
                    <div class="fg-extag">
                      <span class="t-en">Blindfold chess</span>
                      <span class="t-ja">ブラインドフォールド・チェス</span>
                    </div>
                    <div class="fg-exframe" data-example-index="4">
                      <video data-example-video="" data-index="4" src="/assets/fugu_lp/videos/demo4.mp4" controls="" loop="" muted="" playsinline="" preload="metadata" aria-label="Sakana Fugu qualitative example 5"></video>
                      <div data-video-warning="" class="hidden" style="position:absolute;inset:0;flex-direction:column;align-items:center;justify-content:center;gap:6px;padding:0 16px;text-align:center;">
                        <span style="font-size:24px;" aria-hidden="true">⚠️</span>
                        <span style="font-size:14px;font-weight:600;color:rgba(255,255,255,.8);">
                          <span class="t-en">Video unavailable</span>
                          <span class="t-ja">動画を表示できません</span>
                        </span>
                        <span style="font-size:12px;color:rgba(255,255,255,.5);">demo4.mp4</span>
                      </div>
                    </div>
                  </article>
                  <article class="fg-exslide">
                    <div class="fg-extag">
                      <span class="t-en">Time-series · Trading</span>
                      <span class="t-ja">時系列予測 · トレーディング</span>
                    </div>
                    <div class="fg-exframe" data-example-index="5">
                      <video data-example-video="" data-index="5" src="/assets/fugu_lp/videos/demo5.mp4" controls="" loop="" muted="" playsinline="" preload="metadata" aria-label="Sakana Fugu qualitative example 6"></video>
                      <div data-video-warning="" class="hidden" style="position:absolute;inset:0;flex-direction:column;align-items:center;justify-content:center;gap:6px;padding:0 16px;text-align:center;">
                        <span style="font-size:24px;" aria-hidden="true">⚠️</span>
                        <span style="font-size:14px;font-weight:600;color:rgba(255,255,255,.8);">
                          <span class="t-en">Video unavailable</span>
                          <span class="t-ja">動画を表示できません</span>
                        </span>
                        <span style="font-size:12px;color:rgba(255,255,255,.5);">demo5.mp4</span>
                      </div>
                    </div>
                  </article>
                </div>
                <div class="fg-exnav">
                  <button type="button" data-example-prev="" class="fg-exbtn" aria-label="Previous example">
                    <svg width="20" height="20" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m15 18-6-6 6-6"></path>
                    </svg>
                  </button>
                  <div class="fg-exdots">
                    <button type="button" data-example-dot="0" data-active="true" class="fg-exdot" aria-label="Show Example 0"></button>
                    <button type="button" data-example-dot="1" data-active="false" class="fg-exdot" aria-label="Show Example 1"></button>
                    <button type="button" data-example-dot="2" data-active="false" class="fg-exdot" aria-label="Show Example 2"></button>
                    <button type="button" data-example-dot="3" data-active="false" class="fg-exdot" aria-label="Show Example 3"></button>
                    <button type="button" data-example-dot="4" data-active="false" class="fg-exdot" aria-label="Show Example 4"></button>
                    <button type="button" data-example-dot="5" data-active="false" class="fg-exdot" aria-label="Show Example 5"></button>
                  </div>
                  <button type="button" data-example-next="" class="fg-exbtn" aria-label="Next example">
                    <svg width="20" height="20" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                      <path d="m9 18 6-6-6-6"></path>
                    </svg>
                  </button>
                </div>
                <div class="fg-excaps">
                  <div class="fg-excap" data-example-cap="0" data-active="true">
                    <p>
                      <span class="t-en">This experiment shows an AI agent autonomously improving a small GPT's training recipe. Using AutoResearch (Karpathy et al.) – which iteratively edits training code, runs experiments, and keeps only changes that lower validation bits-per-byte (BPB) – the agent ran 123 experiments over ~14 hours on a single H100 GPU. Each line traces a system's best BPB as experiments accumulate: Fugu-Ultra is in bold red (solid = mean over three seeds, dashed = best single run), with three frontier-model baselines (Model A, B, and C) faded behind it, and the callouts mark each new improvement the agent found on its own — spanning batch size, model depth, learning rates, and optimizer settings. Fugu-Ultra finishes with the best mean BPB (0.9774 ± 0.0019), ahead of Model C (0.9781), Model B (0.9793), and Model A (0.9822), and its best single run reaches 0.9748, leading every baseline. This suggests that orchestrating multiple strong models can outperform any individual frontier model on agentic ML research.</span>
                      <span class="t-ja">
                        <strong style="font-weight:600;color:#374151;">例1 — AutoResearch / LLM学習</strong>
                        <br>
                        AIエージェントに小規模なGPTの学習レシピを自律的に改善させる実験。学習コードを反復的に書き換え、実験を実行し、検証用 bits-per-byte（BPB）を下げた変更だけを残していくエージェント型フレームワーク AutoResearch（Karpathy et al.）を用い、エージェントは単一のH100 GPU上でおよそ14時間にわたり123回の実験を実施した。各線は、実験が積み重なるにつれて各システムが達成した最良のBPBの推移を表している。Fugu-Ultra は太い赤の線（実線＝3シードの平均、破線＝最良の単一実行）で示し、その背後に3つのフロンティアモデルのベースライン（Model A・B・C）を淡色で重ねている。吹き出しは、エージェントが自ら見つけた改善点をそれぞれ示しており、バッチサイズ、モデルの深さ、学習率、オプティマイザの設定など多岐にわたる。Fugu-Ultra は最終的に最良の平均BPB（0.9774 ± 0.0019）を達成し、Model C（0.9781）、Model B（0.9793）、Model A（0.9822）を上回った。最良の単一実行では 0.9748 に到達し、すべてのベースラインを上回っている。これらの結果は、複数の強力なモデルをオーケストレーションすることで、エージェント型のML研究において単体のフロンティアモデルを上回り得ることを示唆している。
                      </span>
                    </p>
                  </div>
                  <div class="fg-excap" data-example-cap="1" data-active="false">
                    <p>
                      <span class="t-en">This case study tests whether the reading order of classical Japanese kana letters (仮名消息) can be recovered — letters whose scattered chirashigaki ("scattered-writing") layout makes that genuinely hard even for trained readers of classical Japanese. Each model is given the character bounding boxes together with a rough set of reading-order rules, and writes code that outputs the order the characters should be read in; here it runs on a letter written in 1610 by Hōshun'in (芳春院, 1547–1617), scored by NED (a score based on normalized edit distance from an expert's ground-truth order, where 1.0 is a perfect match). Several frontier models were put through the identical pipeline, but none came close to Fugu-Ultra on this letter: Model A reached only NED 0.24 and Model B scored no better, both far below Fugu-Ultra's 0.80, while Model C produced no predictor at all. The clip shows the two extremes — each panel draws its predicted path in red over the expert's ground truth in green: Fugu-Ultra (top) traces the letter almost exactly, while Model A (bottom) jumps all over the page. (Letter held by the Keio Institute of Oriental Classics.)</span>
                      <span class="t-ja">
                        <strong style="font-weight:600;color:#374151;">例2 — 仮名消息の読み順推定</strong>
                        <br>
                        本ケーススタディは、仮名消息（古典日本語のかな書状）という歴史的資料における読み順の推定問題を対象とする。仮名消息は、文字を紙面に散らして記す「散らし書き」という形式で書かれているため、古文書を読み慣れた人でも文字の読み順を正しく判定することは難しい。そこで各モデルに対して、文字を囲む四角形（バウンディングボックス）と読み順の大まかなルールを与え、文字の読み順を推定するコードを出力させた。実験の対象には1610年に芳春院（ほうしゅんいん、1547–1617）が記した書状を選び、NED（専門家による正しい読み順との正規化編集距離にもとづくスコア。1.0が完全一致）で評価した。複数のフロンティアモデル（A-C）を同一のパイプラインに通したところ、Fugu-Ultraの結果は他のモデルを大きく引き離した。Model AはNED 0.24、Model Bもそれと大差なく、いずれもFugu-Ultraの0.80には遠く及ばない。さらにModel Cはまともなコードを一回も出力できなかった。モデルによる読み順の違いを可視化するために、専門家による正解の読み順（緑）の上に、推定した経路（赤）を描いて映像化した。Fugu-Ultra（上）が読み順をほぼ正確になぞる一方、Model A（下）は紙面全体をあちこち飛び回り、両者は大きく異なる結果を示している。 図：芳春院消息（慶應義塾大学斯道文庫蔵）
                      </span>
                    </p>
                  </div>
                  <div class="fg-excap" data-example-cap="2" data-active="false">
                    <p>
                      <span class="t-en">In this benchmark, each of Fugu-Ultra and 3 frontier models is given a single prompt to write a Rubik's Cube solver from scratch in pure Python — no off-the-shelf solving libraries allowed — and the resulting program is run locally on a held-out set of 300 randomly scrambled cubes. Solution quality is measured by the number of moves a solution uses, where lower is better. Fugu-Ultra and the frontier Model A wrote solvers that ran and solved all 300 cubes, while Model B and Model C each shipped sophisticated-looking code that crashed on execution and returned no valid solution at all (0/300). The clip follows cube #17: from the same scramble, Fugu-Ultra's solver reaches the solved state in 19 moves while Model A needs 21 — and across all 300 cubes Fugu-Ultra averages 19.72 moves versus 19.76 for Model A, both right at the optimal frontier, with Fugu-Ultra never a move longer than Model A on any cube (7 wins, 293 ties, 0 losses).</span>
                      <span class="t-ja">
                        <strong style="font-weight:600;color:#374151;">例3 — ルービックキューブ・ソルバー</strong>
                        <br>
                        本ベンチマークでは、Fugu-Ultraと3つのフロンティアモデルそれぞれに、純粋なPythonのみでルービックキューブソルバーをゼロから実装するよう単一のプロンプトを与えた。既存のソルバーライブラリの使用は禁止とし、生成されたプログラムをランダムにスクランブルされた300個のキューブからなるホールドアウトセットに対してローカルで実行した。解法の質は手数で評価し、少ないほど良いとする。Fugu-Ultraとフロンティアの Model A は300個すべてのキューブを解くソルバーを生成したが、Model B と Model C は一見洗練されたコードを出力したものの、実行時にクラッシュし、有効な解を一つも返せなかった（0/300）。映像はキューブ#17の様子である。同一のスクランブルに対し、Fugu-Ultraのソルバーは19手で完成状態に到達したのに対し、Model A は21手を要した。300個全体の平均では、Fugu-Ultraが19.72手、Model A が19.76手と、いずれも最適解の水準にあり、Fugu-Ultraが Model A より手数が多かったケースは一度もなかった（7勝・293引き分け・0敗）。
                      </span>
                    </p>
                  </div>
                  <div class="fg-excap" data-example-cap="3" data-active="false">
                    <p>
                      <span class="t-en">Task: Create a mechanical iris in CAD, like a camera aperture, where multiple blades move together to open and close the central hole. For each model, we show both the generated detailed CAD itself and a simplified view that makes the structure easier to see. In the CAD generated by Fugu Ultra, the blades rotate around outer pins and clearly open and close the aperture. In contrast, the CAD generated by the other models shows problems such as gaps appearing, weak linkages, or the aperture not closing fully.</span>
                      <span class="t-ja">
                        <strong style="font-weight:600;color:#374151;">例4 — CAD メカニカルアイリス</strong>
                        <br>
                        タスク：カメラの絞り（アパーチャ）のような、複数の羽根が連動して動き中央の穴を開閉する機械式アイリスをCADで作成する。各モデルについて、生成された詳細CAD（Detailed CAD）そのものと、構造を見やすくするための簡易ビュー（Simplified view）の両方を示す。Fugu Ultraが生成したCADでは、羽根が外側のピンを軸に回転し、アパーチャを明確に開閉できている。一方、他のモデルが生成したCADでは、隙間ができてしまう、リンク機構が弱い、アパーチャを十分に閉じきれていない、といった問題が見られる。
                      </span>
                    </p>
                  </div>
                  <div class="fg-excap" data-example-cap="4" data-active="false">
                    <p>
                      <span class="t-en">Four blindfold chess games, back to back. Every model plays the same way — no board shown — holding the full game in memory. Fugu outplays four strong opponents: three leading frontier models and a 2100-Elo Stockfish engine, staying accurate where they drift and ending each game in checkmate.</span>
                      <span class="t-ja">
                        <strong style="font-weight:600;color:#374151;">例5 — 目隠しチェス</strong>
                        <br>
                        4局の目隠しチェスを連続して対局している。すべてのモデルは同じ条件でプレイし、盤面は一切表示されず、ゲーム全体を記憶の中に保持しながら指し手を進める。Fugu は4つの強力な相手——3つの主要なフロンティアモデルと、2100-Elo の Stockfish エンジン——を打ち負かした。相手が手を乱していく場面でも正確さを保ち、いずれの対局もチェックメイトで終えた。
                      </span>
                    </p>
                  </div>
                  <div class="fg-excap" data-example-cap="5" data-active="false">
                    <p>
                      <span class="t-en">This benchmark uses a single anonymized equity over one historical 50-week window and is intended to compare sequential, no-look-ahead decision-making rather than to establish generalizable trading performance. Past performance does not guarantee future results, and results may not transfer to other assets, time periods, or live markets. Each model makes online trading decisions on anonymized STOCK_X, using only current and past weekly market data: opening, high, low, and closing prices, volume, returns, moving averages, volatility, drawdown, portfolio state, and prior feedback. Starting with $10,000, the agent chooses whether to buy, hold, or sell, and what fraction of cash or shares to trade. After each action, the next week's price is revealed and the portfolio is updated, so the model must adapt from feedback rather than seeing the future. Across five runs of the identical 50-week pipeline, Fugu-Ultra grew the portfolio to $11,943.22 ± $633.86, a +19.43% mean return, while the other frontier models reached their return less than +15%.</span>
                      <span class="t-ja">
                        <strong style="font-weight:600;color:#374151;">例6 — 株式トレーディング</strong>
                        <br>
                        匿名化された単一銘柄を1つの過去50週間のウィンドウで用いるこの株式トレーディングのベンチマーク。汎用的なトレーディング性能を立証するためではなく、先読みのない逐次的な意思決定を比較することを目的としている。過去の実績は将来の結果を保証するものではなく、結果が他の資産・期間・実際の市場に当てはまるとは限らない。各モデルは、匿名化された STOCK_X に対して、現在および過去の週次マーケットデータ——始値、高値、安値、終値、出来高、リターン、移動平均、ボラティリティ、ドローダウン、ポートフォリオの状態、直前のフィードバック——のみを用いてオンラインでトレーディングの意思決定を行う。1万ドルからスタートし、エージェントは買い・保有・売りのいずれかと、現金または株式のどの割合を取引するかを選択する。各アクションの後に翌週の価格が開示され、ポートフォリオが更新されるため、モデルは未来を見るのではなくフィードバックから適応しなければならない。同一の50週間パイプラインを5回実行した結果、Fugu-Ultra はポートフォリオを 11,943.22 ± 633.86 ドルまで成長させ、平均リターンは +19.43% に達した。一方、他のフロンティアモデルのリターンはいずれも +15% 未満にとどまった。
                      </span>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
        <!-- §06 FIELD NOTES -->
        <section id="user-voices" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">Users' Voices</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2">
                <span class="t-en">
                  What do our users think about
                  <span class="fg-en">Sakana Fugu</span>
                  ?
                </span>
                <span class="t-ja">
                  <span class="fg-en">Sakana Fugu</span>
                  に対するユーザー評価
                </span>
              </h2>
              <div class="fg-notes" aria-label="User testimonials">
                <article class="fg-note-row">
                  <div class="fg-note-row__n">01</div>
                  <div>
                    <div class="fg-note-row__role">
                      <span class="t-en">Software Engineer</span>
                      <span class="t-ja">ソフトウェアエンジニア</span>
                    </div>
                    <h3 class="fg-note-row__t">
                      <span class="t-en">Coding & Code Review</span>
                      <span class="t-ja">コーディング＆コードレビュー</span>
                    </h3>
                    <blockquote class="fg-note-row__q" style="font-style:normal;">
                      <span class="t-en">
                        For code review, Fugu Ultra is significantly better than GPT-5.5. It gives comprehensive answers and finds the bugs others miss. Where other tools flag about three issues,
                        <span class="fg-en">Sakana Fugu</span>
                        surfaced more than twenty. It's become the model I run all my reviews through.
                      </span>
                      <span class="t-ja">
                        コードレビューでは、Fugu Ultra は回答が網羅的で、他のモデルが見逃すバグまで見つけてくれました。他のツールでは3件くらいの問題しか指摘されなかったが、
                        <span class="fg-en">Sakana Fugu</span>
                        は20件以上を洗い出してくれました。
                      </span>
                    </blockquote>
                  </div>
                </article>
                <article class="fg-note-row">
                  <div class="fg-note-row__n">02</div>
                  <div>
                    <div class="fg-note-row__role">
                      <span class="t-en">Researcher (industry)</span>
                      <span class="t-ja">研究者（企業）</span>
                    </div>
                    <h3 class="fg-note-row__t">
                      <span class="t-en">Research & Autonomy</span>
                      <span class="t-ja">自律的なリサーチ</span>
                    </h3>
                    <blockquote class="fg-note-row__q" style="font-style:normal;">
                      <span class="t-en">I was mapping a patent landscape across ~20 papers and several patents, normally 3–4 days of work. With Fugu I had a full analysis in a few hours, including connections between papers I would never have spotted on my own.</span>
                      <span class="t-ja">
                        約20本の論文と複数の特許にまたがる特許動向（パテントランドスケープ）を作成しました。普段なら3〜4日かかる作業が、
                        <span class="fg-en">Sakana Fugu</span>
                        を使うと数時間で完全な分析ができ、そのなかには、自分では決して気づけなかっただろう論文同士のつながりを見つけることができました。
                      </span>
                    </blockquote>
                  </div>
                </article>
                <article class="fg-note-row">
                  <div class="fg-note-row__n">03</div>
                  <div>
                    <div class="fg-note-row__role">
                      <span class="t-en">Executive (enterprise platform)</span>
                      <span class="t-ja">プラットフォーム企業・役員</span>
                    </div>
                    <h3 class="fg-note-row__t">
                      <span class="t-en">Orchestration</span>
                      <span class="t-ja">オーケストレーション</span>
                    </h3>
                    <blockquote class="fg-note-row__q" style="font-style:normal;">
                      <span class="t-en">Raw output quality is on par with top frontier models, but Fugu showed unusually strong persona stability across long sessions, holding its identity where other models drift. For agent products, that may matter more than raw benchmark scores.</span>
                      <span class="t-ja">
                        素の出力品質はトップクラスのフロンティアモデルと同等。加えて
                        <span class="fg-en">Sakana Fugu</span>
                        は、長時間のセッションでもペルソナが安定しており、他のモデルなら崩れてしまう場面でもキャラクターを保ち続けました。エージェントにとっては、これは単純なベンチマークスコア以上に重要なことです。
                      </span>
                    </blockquote>
                  </div>
                </article>
                <article class="fg-note-row">
                  <div class="fg-note-row__n">04</div>
                  <div>
                    <div class="fg-note-row__role">
                      <span class="t-en">Researcher</span>
                      <span class="t-ja">研究者</span>
                    </div>
                    <h3 class="fg-note-row__t">
                      <span class="t-en">Paper Reproduction</span>
                      <span class="t-ja">論文の再現</span>
                    </h3>
                    <blockquote class="fg-note-row__q" style="font-style:normal;">
                      <span class="t-en">
                        From one simple request,
                        <span class="fg-en">Sakana Fugu</span>
                        worked autonomously for nearly four hours — reading the paper, implementing, training, evaluating, and analyzing the gaps.
                      </span>
                      <span class="t-ja">
                        一つのシンプルな指示から、
                        <span class="fg-en">Sakana Fugu</span>
                        はおよそ4時間続けて自律的に作業しました。論文を読み込み、実装・学習・評価まで行い、足りない点を分析してくれました。
                      </span>
                    </blockquote>
                  </div>
                </article>
                <article class="fg-note-row">
                  <div class="fg-note-row__n">05</div>
                  <div>
                    <div class="fg-note-row__role">
                      <span class="t-en">Security Engineer</span>
                      <span class="t-ja">セキュリティエンジニア</span>
                    </div>
                    <h3 class="fg-note-row__t">
                      <span class="t-en">Security Assessment</span>
                      <span class="t-ja">セキュリティ評価</span>
                    </h3>
                    <blockquote class="fg-note-row__q" style="font-style:normal;">
                      <span class="t-en">
                        Given one scoped instruction,
                        <span class="fg-en">Sakana Fugu</span>
                        drove a full security assessment end-to-end — recon, XSS/SQLi checks, auth review, and a clean report with evidence and retest steps — staying inside scope and avoiding destructive actions.
                      </span>
                      <span class="t-ja">
                        範囲を絞った指示を一つ渡しただけで、
                        <span class="fg-en">Sakana Fugu</span>
                        は情報収集から XSS/SQLi の検査、認証まわりのレビュー、さらに証拠と再テスト手順を備えた整然としたレポート作成まで、セキュリティ評価を一気通貫でこなしました。しかも指定した範囲を逸脱せず、システムを壊すような操作も避けてくれました。
                      </span>
                    </blockquote>
                  </div>
                </article>
              </div>
            </div>
          </div>
        </section>
        <!-- §07 PRICING -->
        <section id="pricing" class="fg-sec fg-sec--alt">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">Pricing</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 class="fg-h2" style="margin-bottom:clamp(28px,3vw,40px);">
                <span class="t-en">Pricing Plan</span>
                <span class="t-ja">料金プラン</span>
              </h2>
              <div aria-label="Sakana Fugu pricing plans">
                <article class="fg-plan">
                  <div class="fg-plan__head">
                    <div>
                      <span class="fg-mono" style="font-size:12px;color:#9ca3af;">01</span>
                      <span class="fg-plan__title">
                        <span class="t-en">Token Plan</span>
                        <span class="t-ja">トークンプラン</span>
                      </span>
                    </div>
                    <span class="fg-tag fg-tag--ac">
                      <span class="t-en">Pay-as-you-go</span>
                      <span class="t-ja">従量課金制</span>
                    </span>
                  </div>
                  <p class="fg-plan__desc">
                    <span class="t-en">For heavy, production workloads needing maximum reliability — consumption-based tokens are served at higher priority than monthly-plan tokens.</span>
                    <span class="t-ja">最大限の信頼性が求められる高負荷・本番ワークロード向け。従量課金のトークンは、月額プランのトークンより高い優先度で処理されます。</span>
                  </p>
                  <div class="fg-subcells fg-subcells--stack">
                    <div class="fg-subcell">
                      <div class="fg-subcell__name">Fugu</div>
                      <div class="fg-subcell__lbl">
                        <span class="t-en">When 1 agent is active</span>
                        <span class="t-ja">エージェントが 1 つの場合</span>
                      </div>
                      <p class="fg-subcell__p">
                        <span class="t-en">You pay only the standard rate for that specific underlying model.</span>
                        <span class="t-ja">その基盤モデルの標準レートのみをお支払いいただきます。</span>
                      </p>
                      <div class="fg-subcell__lbl">
                        <span class="t-en">When multiple agents are active</span>
                        <span class="t-ja">複数のエージェントが稼働している場合</span>
                      </div>
                      <p class="fg-subcell__p">
                        <span class="t-en">We never stack model fees; you are charged a single rate based on the top tier model involved.</span>
                        <span class="t-ja">モデル料金を積み上げることはありません。関与する最上位モデルに基づく単一のレートで課金されます。</span>
                      </p>
                    </div>
                    <div class="fg-subcell">
                      <div class="fg-subcell__name">Fugu Ultra</div>
                      <div class="fg-subcell__sub">
                        <span class="t-en">Fixed pricing for fugu-ultra-v2.0</span>
                        <span class="t-ja">fugu-ultra-v2.0 の料金（一律）</span>
                      </div>
                      <div class="fg-pricegrid">
                        <div class="fg-pricecell">
                          <div class="fg-pricecell__l">
                            <span class="t-en">Input</span>
                            <span class="t-ja">入力</span>
                          </div>
                          <div class="fg-pricecell__p">$5
                            <span>
                              <span class="t-en">/1M tokens</span>
                              <span class="t-ja">/100万トークン</span>
                            </span>
                          </div>
                          <div class="fg-pricecell__n">
                            <span class="t-en">$10 when context > 272K</span>
                            <span class="t-ja">$10（コンテキスト272K超）</span>
                          </div>
                        </div>
                        <div class="fg-pricecell">
                          <div class="fg-pricecell__l">
                            <span class="t-en">Output</span>
                            <span class="t-ja">出力</span>
                          </div>
                          <div class="fg-pricecell__p">$30
                            <span>
                              <span class="t-en">/1M tokens</span>
                              <span class="t-ja">/100万トークン</span>
                            </span>
                          </div>
                          <div class="fg-pricecell__n">
                            <span class="t-en">$45 when context > 272K</span>
                            <span class="t-ja">$45（コンテキスト272K超）</span>
                          </div>
                        </div>
                        <div class="fg-pricecell">
                          <div class="fg-pricecell__l">
                            <span class="t-en">Cached input</span>
                            <span class="t-ja">キャッシュ入力</span>
                          </div>
                          <div class="fg-pricecell__p">$0.50
                            <span>
                              <span class="t-en">/1M tokens</span>
                              <span class="t-ja">/100万トークン</span>
                            </span>
                          </div>
                          <div class="fg-pricecell__n">
                            <span class="t-en">$1.00 when context > 272K</span>
                            <span class="t-ja">$1.00（コンテキスト272K超）</span>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="fg-subcell">
                      <div class="fg-subcell__name">Fugu Max</div>
                      <div class="fg-subcell__sub">
                        <span class="t-en">Fixed rates regardless of context length. web_search and web_fetch are billed at $0.007 per call.</span>
                        <span class="t-ja">コンテキスト長にかかわらず一律料金。web_search と web_fetch は 1 回あたり $0.007 で課金されます。</span>
                      </div>
                      <div class="fg-pricegrid">
                        <div class="fg-pricecell">
                          <div class="fg-pricecell__l">
                            <span class="t-en">Input</span>
                            <span class="t-ja">入力</span>
                          </div>
                          <div class="fg-pricecell__p">$2</div>
                        </div>
                        <div class="fg-pricecell">
                          <div class="fg-pricecell__l">
                            <span class="t-en">Output</span>
                            <span class="t-ja">出力</span>
                          </div>
                          <div class="fg-pricecell__p">$6</div>
                        </div>
                        <div class="fg-pricecell">
                          <div class="fg-pricecell__l">
                            <span class="t-en">Cached input</span>
                            <span class="t-ja">キャッシュ入力</span>
                          </div>
                          <div class="fg-pricecell__p">$0.25</div>
                        </div>
                      </div>
                    </div>
                    <div class="fg-subcell">
                      <div class="fg-subcell__name">Fugu Cyber</div>
                      <div class="fg-subcell__sub">
                        <span class="t-en">Please contact our sales team for details on Fugu Cyber usage and pricing.</span>
                        <span class="t-ja">Fugu Cyberのご利用や料金の詳細につきましては、営業チームにお問い合わせください。</span>
                      </div>
                      <div style="margin-top:16px;">
                        <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark">
                          <span class="t-en">Contact Us</span>
                          <span class="t-ja">お問い合わせ</span>
                        </a>
                      </div>
                    </div>
                  </div>
                </article>
                <article class="fg-plan">
                  <div class="fg-plan__head">
                    <div>
                      <span class="fg-mono" style="font-size:12px;color:#9ca3af;">02</span>
                      <span class="fg-plan__title">
                        <span class="t-en">Subscription Plan</span>
                        <span class="t-ja">サブスクリプションプラン</span>
                      </span>
                    </div>
                    <span class="fg-tag fg-tag">
                      <span class="t-en">Monthly</span>
                      <span class="t-ja">月額</span>
                    </span>
                  </div>
                  <p class="fg-plan__desc">
                    <span class="t-en">
                      We recommend the <strong style="font-weight:600;color:#111827;">Token Plan</strong> for serious workloads. We offer subscription plans for casual, everyday hands-on use. Every tier includes
                      <strong style="font-weight:600;color:#111827;">Fugu</strong>,
                      <strong style="font-weight:600;color:#111827;">Fugu Ultra</strong>, and
                      <strong style="font-weight:600;color:#111827;">Fugu Max</strong>
                      — upgrade when you need longer, heavier, or more frequent sessions.
                    </span>
                    <span class="t-ja">
                      本格的なワークロードには<strong style="font-weight:600;color:#111827;">トークンプラン</strong>をお勧めします。日常的なカジュアルな用途にはサブスクリプションプランをご用意しています。すべてのプランで
                      <strong style="font-weight:600;color:#111827;">Fugu</strong>、
                      <strong style="font-weight:600;color:#111827;">Fugu Ultra</strong>、
                      <strong style="font-weight:600;color:#111827;">Fugu Max</strong>
                      をご利用いただけます。より長時間・高負荷・高頻度の作業が必要な場合は上位プランへ。
                    </span>
                  </p>
                  <div class="fg-tiers" aria-label="Subscription tiers">
                    <div class="fg-tier">
                      <div class="fg-tier__name">Standard Plan</div>
                      <div class="fg-tier__price">
                        $20
                        <span>
                          <span class="t-en">/month</span>
                          <span class="t-ja">/月</span>
                        </span>
                      </div>
                      <div class="fg-tier__sub">
                        <span class="t-en">Lightweight daily usage</span>
                        <span class="t-ja">軽量な日常利用に</span>
                      </div>
                      <div class="fg-tier__desc">
                        <span class="t-en">For occasional API calls, small experiments, and trying Fugu in personal workflows.</span>
                        <span class="t-ja">低頻度の API 利用、小規模な実験、個人ワークフローでの試用に。</span>
                      </div>
                      <div class="fg-tier__badge">
                        <span class="t-en">Baseline allowance</span>
                        <span class="t-ja">標準の利用枠</span>
                      </div>
                    </div>
                    <div class="fg-tier">
                      <div class="fg-tier__name">Pro Plan</div>
                      <div class="fg-tier__price">
                        $100
                        <span>
                          <span class="t-en">/month</span>
                          <span class="t-ja">/月</span>
                        </span>
                      </div>
                      <div class="fg-tier__sub">
                        <span class="t-en">Focused working sessions</span>
                        <span class="t-ja">集中した作業セッションに</span>
                      </div>
                      <div class="fg-tier__desc">
                        <span class="t-en">For regular coding, review, research, and analysis sessions throughout the week.</span>
                        <span class="t-ja">普段のコーディング、レビュー、調査、分析セッションに。</span>
                      </div>
                      <div class="fg-tier__badge">
                        <span class="t-en">10× Standard Plan usage</span>
                        <span class="t-ja">Standard Plan の 10 倍の利用枠</span>
                      </div>
                    </div>
                    <div class="fg-tier">
                      <div class="fg-tier__name">Max Plan</div>
                      <div class="fg-tier__price">
                        $200
                        <span>
                          <span class="t-en">/month</span>
                          <span class="t-ja">/月</span>
                        </span>
                      </div>
                      <div class="fg-tier__sub">
                        <span class="t-en">Heavy long-running workloads</span>
                        <span class="t-ja">長時間の高負荷ワークロードに</span>
                      </div>
                      <div class="fg-tier__desc">
                        <span class="t-en">For power users who keep Fugu active across deeper, longer-running tasks.</span>
                        <span class="t-ja">
                          より深く長時間のタスクで
                          <span class="fg-en">Sakana Fugu</span>
                          を継続的に使うパワーユーザー向け。
                        </span>
                      </div>
                      <div class="fg-tier__badge">
                        <span class="t-en">20× Standard Plan usage</span>
                        <span class="t-ja">Standard Plan の 20 倍の利用枠</span>
                      </div>
                    </div>
                  </div>
                  <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark" style="margin-top:24px;">
                    <span class="t-en">
                      Start Using
                      <span class="fg-en">Sakana Fugu</span>
                    </span>
                    <span class="t-ja">今すぐはじめる</span>
                  </a>
                </article>
              </div>
            </div>
          </div>
        </section>
        <!-- §08 FAQ -->
        <section id="faq" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">FAQ</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div class="fg-faqs" aria-label="Frequently asked questions">
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-1">
                  <span class="fg-faq__n">Q1</span>
                  <span class="fg-faq__q">
                    <span class="t-en">
                      How do I get started with
                      <span class="fg-en">Sakana Fugu</span>
                      ?
                    </span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      はどうやって使い始めますか？
                    </span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-1" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">
                      <span class="fg-en">Sakana Fugu</span>
                      is available through an OpenAI-compatible API. Point your existing client or coding harness at the Fugu endpoint with your API key and start sending requests — no SDK migration required.
                    </span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      は OpenAI 互換 API を通じて利用できます。既存のクライアントやコーディングハーネスを、API キーとともに Fugu のエンドポイントに向けてリクエストを送るだけです——SDK の移行は必要ありません。
                    </span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-2">
                  <span class="fg-faq__n">Q2</span>
                  <span class="fg-faq__q">
                    <span class="t-en">What's the difference between Fugu, Fugu Ultra, and Fugu Max?</span>
                    <span class="t-ja">Fugu、Fugu Ultra、Fugu Max の違いは何ですか？</span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-2" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">Fugu balances latency and quality, making it a strong default for everyday coding and interactive work. Fugu Ultra prioritizes answer quality on complex, multi-step reasoning, coordinating more expert agents when accuracy and depth matter most, at the cost of response time. Fugu Max orchestrates our largest pool of models and is optimized for the overall cost–performance tradeoff, dynamically finding efficient combinations of expert agents to improve model performance and cost efficiency together.</span>
                    <span class="t-ja">Fugu はレイテンシと品質のバランスが取れており、日常的なコーディングやインタラクティブな作業に適した標準モデルです。Fugu Ultra は、複雑で多段階の推論において回答品質を最優先し、精度と深さが重要な場面ではより多くの専門エージェントを連携させます（その分、応答時間は長くなります）。Fugu Max は最大規模のモデルプールをオーケストレーションし、コストと性能の総合的なバランスに最適化されています。専門エージェントを効率的に組み合わせることで、モデル性能とコスト効率の両方を高めることを目指します。</span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-3">
                  <span class="fg-faq__n">Q3</span>
                  <span class="fg-faq__q">
                    <span class="t-en">Can I control which models or providers Fugu uses?</span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      が使用するモデルやプロバイダーを制御できますか？
                    </span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-3" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">
                      Fugu Ultra relies on the full agent pool to deliver its performance, so its pool is fixed. Fugu Max must consider both cost and performance when optimizing its orchestration, so its pool is also fixed. For Fugu, you can opt out of specific models from the settings menu on our console page to match your data, privacy, and compliance needs.
                      <strong style="font-weight:600;color:#111827;">We also provide customized services for enterprise customers who require specific model or provider configurations, so please contact us to discuss such requests.</strong>
                    </span>
                    <span class="t-ja">
                      Fugu Ultra はその性能を発揮するために全エージェントプールを利用するため、プールは固定です。Fugu Max も、コストと性能の両方を考慮してオーケストレーションを最適化する必要があるため、プールは固定です。Fugu については、コンソールページの設定メニューから特定のモデルを除外でき、データ・プライバシー・コンプライアンスの要件に合わせられます。
                      <strong style="font-weight:600;color:#111827;">また、特定のモデルやプロバイダー構成を必要とするエンタープライズのお客様には個別対応サービスをご提供していますので、こうしたご要望についてはぜひお問い合わせください。</strong>
                    </span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-4">
                  <span class="fg-faq__n">Q4</span>
                  <span class="fg-faq__q">
                    <span class="t-en">How often do Fugu models update to add new frontier models?</span>
                    <span class="t-ja">
                      <span class="fg-en">Sakana Fugu</span>
                      モデルはどのくらいの頻度で新しいフロンティアモデルを追加・更新しますか？
                    </span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-4" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">We aim to give users the best performance available. When a new frontier model is released publicly, we expect to spend roughly two weeks training and evaluating updated Fugu models before rolling them out.</span>
                    <span class="t-ja">
                      私たちは利用可能な最高の性能をユーザーに提供することを目指しています。新しいフロンティアモデルが一般公開された場合、アップデート版の
                      <span class="fg-en">Sakana Fugu</span>
                      モデルのトレーニングと評価に約 2 週間をかけ、その後順次提供開始していく予定です。
                    </span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-5">
                  <span class="fg-faq__n">Q5</span>
                  <span class="fg-faq__q">
                    <span class="t-en">How does pricing work?</span>
                    <span class="t-ja">料金はどのように決まりますか？</span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-5" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">We offer both subscription and pay-as-you-go plans, and every plan includes access to Fugu, Fugu Ultra, and Fugu Max. The subscription plan has three monthly tiers: Standard Plan ($20/month) is great for lightweight daily use; Pro Plan ($100/month) provides 10× the usage of Standard Plan, ideal for a few focused working sessions each week; and Max Plan ($200/month) provides 20× the usage of Standard Plan, built for heavy, long-running workloads. The pay-as-you-go plan bills by token usage instead of a monthly allowance, giving you elastic capacity for spikes and large jobs—ideal for enterprise customers. With it, Fugu is charged at the standard rate of the underlying model, and when multiple agents are active, we never stack fees: you pay a single rate based on the top-tier model involved. Fugu Ultra (fugu-ultra-v2.0) is priced per 1M tokens at $5 input, $30 output, and $0.50 cached input, with higher rates ($10 / $45 / $1.00) for contexts above 272K tokens. Fugu Max is priced per 1M tokens at $2 input, $6 output, and $0.25 cached input regardless of context length; web_search and web_fetch are billed at $0.007 per call.</span>
                    <span class="t-ja">サブスクリプションプランと従量課金プランの両方をご用意しており、すべてのプランで Fugu、Fugu Ultra、Fugu Max をご利用いただけます。サブスクリプションプランには月額 3 つのプランがあります。Standard Plan（$20/月）は軽量な日常利用に最適、Pro Plan（$100/月）は Standard Plan の 10 倍の利用量を提供し週に数回の集中作業に向いており、Max Plan（$200/月）は Standard Plan の 20 倍の利用量を提供し長時間にわたる負荷の高いワークロード向けです。従量課金プランは月額の利用枠ではなくトークン使用量に応じて課金され、スパイクや大規模ジョブに対する柔軟なキャパシティを提供します。エンタープライズのお客様にも最適です。このプランでは、Fugu は基盤モデルの標準レートで課金され、複数のエージェントが稼働している場合でも料金を積み上げることはありません。関与する最上位モデルに基づく単一のレートでお支払いいただきます。Fugu Ultra（fugu-ultra-v2.0）は 100 万トークンあたり、入力 $5、出力 $30、キャッシュ入力 $0.50 で、272K トークンを超えるコンテキストではより高いレート（$10 / $45 / $1.00）が適用されます。Fugu Max はコンテキスト長にかかわらず、100 万トークンあたり入力 $2、出力 $6、キャッシュ入力 $0.25 で、web_search と web_fetch は 1 回あたり $0.007 で課金されます。</span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-6">
                  <span class="fg-faq__n">Q6</span>
                  <span class="fg-faq__q">
                    <span class="t-en">The pricing for Fugu with opt-out agents sounds complex. Can you give an example?</span>
                    <span class="t-ja">
                      エージェントを除外できる
                      <span class="fg-en">Sakana Fugu</span>
                      の料金は複雑に思えます。例を挙げてもらえますか？
                    </span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-6" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">Yes. Think of Fugu pricing as a single blended rate for the active agent pool, not a sum of every model used. If your pool contains only Model A, requests are billed at Model A's rate. If your pool contains Models A, B, and C, you still pay only one rate: the rate of the top tier model among A, B, and C. In other words, adding more agents does not multiply the bill; it only determines which single model rate applies to that configured pool.</span>
                    <span class="t-ja">
                      はい。
                      <span class="fg-en">Sakana Fugu</span>
                      の料金は、使用したすべてのモデルの合計ではなく、稼働中のエージェントプールに対する単一のブレンドレートだと考えてください。プールにモデル A だけが含まれている場合、リクエストはモデル A のレートで課金されます。プールにモデル A・B・C が含まれている場合でも、お支払いは 1 つのレート——A・B・C のうち最上位モデルのレート——だけです。つまり、エージェントを増やしても請求が積算されるわけではなく、その構成済みプールに適用される単一のモデルレートが決まるだけです。
                    </span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-7">
                  <span class="fg-faq__n">Q7</span>
                  <span class="fg-faq__q">
                    <span class="t-en">Can I see and predict my token usage and cost?</span>
                    <span class="t-ja">トークン使用量とコストを確認・予測できますか？</span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-7" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">Yes. Token usage and the corresponding cost are reported per request, so you can monitor spend in real time and forecast costs before scaling up.</span>
                    <span class="t-ja">はい。トークン使用量と対応するコストはリクエストごとに報告されるため、支出をリアルタイムで把握し、スケールアップ前にコストを予測することができます。</span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-8">
                  <span class="fg-faq__n">Q8</span>
                  <span class="fg-faq__q">
                    <span class="t-en">Will my data be used to train models? Can I opt out?</span>
                    <span class="t-ja">自分のデータはモデルの学習に使われますか？オプトアウトできますか？</span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-8" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">Usage data helps us keep improving Fugu's performance, and we're grateful when customers share it. That said, it's entirely your choice — you can opt out of training data usage at any time from our console page.</span>
                    <span class="t-ja">
                      学習データ利用についてはコンソールページからいつでもオプトアウトすることができます。お客様のご判断でご共有いただける場合は
                      <span class="fg-en">Sakana Fugu</span>
                      の性能向上に役立てます。
                    </span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-9">
                  <span class="fg-faq__n">Q9</span>
                  <span class="fg-faq__q">
                    <span class="t-en">Can I see which underlying models Fugu used for each query?</span>
                    <span class="t-ja">
                      各クエリで
                      <span class="fg-en">Sakana Fugu</span>
                      が使用した基盤モデルを確認できますか？
                    </span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-9" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">No. The specific models Fugu selects and how it coordinates them are proprietary, so this routing information is not exposed by design.</span>
                    <span class="t-ja">
                      いいえ。
                      <span class="fg-en">Sakana Fugu</span>
                      が選択する具体的なモデルやそれらをどのように連携させるかは独自技術であり、この設計情報は公開していません。
                    </span>
                  </p>
                </div>
              </article>
              <article data-faq-item="" data-open="false" class="fg-faq">
                <button type="button" data-faq-button="" class="fg-faq__btn" aria-expanded="false" aria-controls="faq-answer-10">
                  <span class="fg-faq__n">Q10</span>
                  <span class="fg-faq__q">
                    <span class="t-en">
                      Can I use
                      <span class="fg-en">Sakana Fugu</span>
                      from outside Japan?
                    </span>
                    <span class="t-ja">海外からも利用できますか？</span>
                  </span>
                  <span class="fg-faq__icon" aria-hidden="true">
                    <svg class="fg-plus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                      <path d="M12 5v14"></path>
                    </svg>
                    <svg class="fg-minus" width="18" height="18" viewbox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                      <path d="M5 12h14"></path>
                    </svg>
                  </span>
                </button>
                <div id="faq-answer-10" data-faq-answer="" class="fg-faq__ans">
                  <p data-open="false">
                    <span class="t-en">Yes, Fugu is available from outside Japan. However, we do not provide services to users in EU (European Union) or EEA (European Economic Area) member states (please refer to our Terms of Service for details). Additionally, in other regions, access may not be available due to network conditions or local regulations.</span>
                    <span class="t-ja">はい、日本国外からもご利用いただけます。ただし、EU（欧州連合）およびEEA（欧州経済領域）加盟国へのサービス提供は行っておりません（詳細は利用規約をご確認ください）。また、それ以外の地域におきましても、通信環境や現地の各種規制等によってご利用いただけない場合がございます。</span>
                  </p>
                </div>
              </article>
            </div>
          </div>
        </section>
        <!-- §09 CONTACT -->
        <section id="contact" class="fg-sec">
          <div class="fg-grid">
            <div class="fg-aside">
              <div class="fg-aside__label">Contact Us</div>
              <div class="fg-aside__bar"></div>
            </div>
            <div>
              <h2 style="margin:0;font-size:clamp(32px,4.4vw,60px);font-weight:700;line-height:1.18;letter-spacing:-.01em;color:#111827;">
                <span class="t-en">
                  Ready to build with
                  <span style="color:var(--ac);">
                    <span class="fg-en">Sakana Fugu</span>
                  </span>
                  ?
                </span>
                <span class="t-ja">
                  <span style="color:var(--ac);">
                    <span class="fg-en">Sakana Fugu</span>
                  </span>
                  を使った開発を始めてみませんか？
                </span>
              </h2>
              <p class="fg-lead" style="max-width:560px;">
                <span class="t-en">Get in touch to learn more about access, plans, and enterprise deployment.</span>
                <span class="t-ja">アクセス方法やプラン、エンタープライズ向けの導入に関する詳細については、お問い合わせください。</span>
              </p>
              <div style="display:flex;flex-direction:column;align-items:flex-start;gap:20px;margin-top:32px;">
                <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-btn fg-btn--dark" style="height:54px;padding:0 32px;font-size:15px;">
                  <span class="t-en">Contact Us</span>
                  <span class="t-ja">お問い合わせ</span>
                </a>
              </div>
            </div>
          </div>
        </section>
      </main>
      <footer class="fg-footer">
        <div class="fg-footer__inner">
          <div class="fg-footer__logo">
            <img src="/assets/fugu_lp/Logo_SakanaFugu.png" alt="Sakana Fugu">
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
          <div class="fg-footer__copy">
            <span class="t-en">© 2026 Sakana AI 株式会社</span>
            <span class="t-ja">© 2026 Sakana AI 株式会社</span>
          </div>
        </div>
      </footer>
      <div class="fg-mcta">
        <a href="https://console.sakana.ai/login" target="_blank" rel="noreferrer" class="fg-mcta__btn">
          <span class="t-en">
            Start Using
            <span class="fg-en">Sakana Fugu</span>
          </span>
          <span class="t-ja">今すぐはじめる</span>
        </a>
      </div>
    </div>
    <script src="/js/fugu-lp-init.js"></script>
    <script data-fugu-inline-carousel-fallback="">
(function () {
  function closest(el, selector) {
    while (el && el.nodeType === 1) {
      if (el.matches && el.matches(selector)) return el;
      el = el.parentElement;
    }
    return null;
  }

  function getSlides(track) {
    return Array.prototype.filter.call(track.children, function (child) {
      return child && child.tagName && child.tagName.toLowerCase() === "article";
    });
  }

  function currentIndex(track, slides) {
    if (!slides.length) return 0;
    var best = 0;
    var bestDist = Infinity;
    for (var i = 0; i < slides.length; i++) {
      var left = slides[i].offsetLeft - track.offsetLeft;
      var dist = Math.abs(left - track.scrollLeft);
      if (dist < bestDist) {
        best = i;
        bestDist = dist;
      }
    }
    return best;
  }

  function setDots(index) {
    var dots = document.querySelectorAll('[data-example-dot]');
    for (var i = 0; i < dots.length; i++) {
      var active = String(i) === String(index);
      dots[i].setAttribute('data-active', active ? 'true' : 'false');
      if (active) dots[i].setAttribute('aria-current', 'true');
      else dots[i].removeAttribute('aria-current');
    }
  }

  function go(deltaOrIndex, absolute) {
    var track = document.querySelector('[data-examples-track]');
    if (!track) return;
    var slides = getSlides(track);
    if (!slides.length) return;
    var index = absolute ? deltaOrIndex : currentIndex(track, slides) + deltaOrIndex;
    index = (index + slides.length) % slides.length;
    var left = slides[index].offsetLeft - track.offsetLeft;
    if (typeof track.scrollTo === 'function') {
      track.scrollTo({ left: left, behavior: 'smooth' });
    } else {
      track.scrollLeft = left;
    }
    setDots(index);
  }

  document.addEventListener('click', function (event) {
    var prev = closest(event.target, '[data-example-prev]');
    var next = closest(event.target, '[data-example-next]');
    var dot = closest(event.target, '[data-example-dot]');
    if (!prev && !next && !dot) return;
    event.preventDefault();
    event.stopPropagation();
    if (event.stopImmediatePropagation) event.stopImmediatePropagation();
    if (prev) go(-1, false);
    else if (next) go(1, false);
    else if (dot) go(parseInt(dot.getAttribute('data-example-dot') || '0', 10), true);
  }, true);

  var syncTimer;
  document.addEventListener('scroll', function () {}, true);
  window.addEventListener('load', function () {
    var track = document.querySelector('[data-examples-track]');
    if (!track) return;
    var slides = getSlides(track);
    setDots(currentIndex(track, slides));
    track.addEventListener('scroll', function () {
      window.clearTimeout(syncTimer);
      syncTimer = window.setTimeout(function () {
        setDots(currentIndex(track, getSlides(track)));
      }, 80);
    });
  });
})();
    </script>
    <script data-fugu-partners-strip="">
(function () {
  // Availability / partners strip: pick the layout from the MEASURED logo width,
  // never the logo count, so adding partners needs no code change.
  //   fits beside label & <6 logos → "inline"  (static centered row)
  //   ≥6 logos, or fits only on its own row → "stacked" (label on top)
  //   overflows a narrow strip (≤560px) with ≤6 logos → "grid" (2 columns)
  //   any other overflow → "marquee" (endless scroll)
  var SPEED = 40;      // marquee travel speed, px/s
  var GRID_MAX = 6;    // ≤ this many logos may fall back to the 2-col grid
  var NARROW = 560;    // strip width (px) at/below which overflow grids, else marquees

  var section = document.querySelector('[data-partners-strip]');
  if (!section) return;
  var inner = section.querySelector('.fg-partners__inner');
  var label = section.querySelector('.fg-partners__label');
  var vp = section.querySelector('.fg-partners__vp');
  var track = section.querySelector('.fg-partners__track');
  var set = section.querySelector('.fg-partners__logos');
  if (!inner || !label || !vp || !track || !set) return;

  // A non-interactive duplicate set trailing the original makes the loop seamless.
  var clone = set.cloneNode(true);
  clone.classList.add('fg-partners__logos--clone');
  clone.setAttribute('aria-hidden', 'true');
  Array.prototype.forEach.call(clone.querySelectorAll('a'), function (a) { a.setAttribute('tabindex', '-1'); });
  Array.prototype.forEach.call(clone.querySelectorAll('img'), function (img) { img.setAttribute('alt', ''); });
  track.appendChild(clone);

  function px(v) { return parseFloat(v) || 0; }

  var lastSig = '';
  function update() {
    var kids = set.children;
    var n = kids.length;
    if (!n) return;

    // logo widths are gap-independent, so measure them directly (stays correct
    // even once the set is display:grid, since grid items are content-sized)
    var gap = px(getComputedStyle(set).columnGap) || px(getComputedStyle(set).gap);
    var content = 0;
    for (var i = 0; i < n; i++) content += kids[i].getBoundingClientRect().width;
    var need = content + gap * (n - 1);

    var innerCS = getComputedStyle(inner);
    var innerW = inner.clientWidth - px(innerCS.paddingLeft) - px(innerCS.paddingRight);
    var innerGap = px(innerCS.columnGap) || px(innerCS.gap);
    var availInline = innerW - label.getBoundingClientRect().width - innerGap; // logos beside label
    var availStacked = innerW;                                                 // logos on their own row
    var narrow = section.getBoundingClientRect().width <= NARROW;

    var mode;
    if (need <= availInline + 1 && n < GRID_MAX) mode = 'inline';
    else if (need <= availStacked + 1) mode = 'stacked';
    else mode = (narrow && n <= GRID_MAX) ? 'grid' : 'marquee';

    var dur = ((need + gap) / SPEED).toFixed(2);
    // signature guards against re-writing identical styles → ResizeObserver can't loop
    var sig = mode + '|' + n + '|' + Math.round(need) + '|' + Math.round(availInline) + '|' + Math.round(availStacked) + '|' + dur;
    if (sig === lastSig) return;
    lastSig = sig;

    section.setAttribute('data-mode', mode);

    // odd logo out in a 2-col grid → span both columns so it centers on its row
    for (var j = 0; j < n; j++) kids[j].style.gridColumn = '';
    if (mode === 'grid' && n % 2 === 1) kids[n - 1].style.gridColumn = '1 / -1';

    if (mode === 'marquee') track.style.setProperty('--fg-marquee-dur', dur + 's');
    else track.style.removeProperty('--fg-marquee-dur');
  }

  var raf = 0;
  function schedule() {
    if (raf) cancelAnimationFrame(raf);
    raf = requestAnimationFrame(update);
  }

  if (window.ResizeObserver) {
    var ro = new ResizeObserver(schedule);
    ro.observe(inner);
    ro.observe(set);
  }
  window.addEventListener('resize', schedule);
  if (document.fonts && document.fonts.ready) document.fonts.ready.then(schedule).catch(function () {});
  // logos may decode after first paint — recompute as each lands
  Array.prototype.forEach.call(section.querySelectorAll('img'), function (img) {
    if (!img.complete) img.addEventListener('load', schedule, { once: true });
  });
  update();
})();
    </script>
  </body>
</html>
