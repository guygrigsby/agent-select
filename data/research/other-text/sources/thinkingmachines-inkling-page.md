<!DOCTYPE html><html lang="en" class="no-js"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><script>
      document.documentElement.classList.remove('no-js');
      document.documentElement.classList.add('js');
    </script><meta name="author" content="Thinking Machines Lab"><meta name="description" content="Efficient open-weights models that reason over text, image, and audio. Fine-tune them on Tinker."><link rel="apple-touch-icon" sizes="180x180" href="/images/apple-touch-icon.png"><link rel="icon" type="image/png" sizes="32x32" href="/images/favicon-32x32.png"><link rel="icon" type="image/png" sizes="16x16" href="/images/favicon-16x16.png"><link rel="preload" href="/fonts/iowan-old-style/iowan-old-style_regular.woff2" as="font" type="font/woff2" crossorigin="anonymous" fetchpriority="high"><link rel="preload" href="/fonts/gt-america/GT-America-Standard-Regular.woff2" as="font" type="font/woff2" crossorigin="anonymous"><link rel="preload" href="/fonts/gt-america/GT-America-Standard-Medium.woff2" as="font" type="font/woff2" crossorigin="anonymous"><link rel="preload" href="/fonts/chakra-petch/chakra-petch-v12-latin-500.woff2" as="font" type="font/woff2" crossorigin="anonymous"><meta name="keywords" content="thinky thinkingmachines machine learning deep learning ai"><link rel="canonical" href="https://thinkingmachines.ai/inkling/"><meta itemprop="name" content="Inkling"><meta itemprop="description" content="Efficient open-weights models that reason over text, image, and audio. Fine-tune them on Tinker."><meta itemprop="datePublished" content="2026-07-07T00:00:00+00:00"><meta itemprop="dateModified" content="2026-08-31T11:19:20-07:00"><link rel="stylesheet" href="/css/common.css?v=c32b5950996e" fetchpriority="high"><link rel="stylesheet" href="/css/ui.css?v=7c58697c7635"><link rel="preload" as="style" href="/css/typography.css?v=fbb994bc49ca"><link rel="stylesheet" href="/css/typography.css?v=fbb994bc49ca" media="print" onload="this.media='all'"><noscript><link rel="stylesheet" href="/css/typography.css?v=fbb994bc49ca"></noscript><link rel="preload" as="style" href="/css/fonts.css?v=446f09d40630"><link rel="stylesheet" href="/css/fonts.css?v=446f09d40630" media="print" onload="this.media='all'"><noscript><link rel="stylesheet" href="/css/fonts.css?v=446f09d40630"></noscript><link rel="preload" as="style" href="/css/syntax.css?v=3da135297e59"><link rel="stylesheet" href="/css/syntax.css?v=3da135297e59" media="print" onload="this.media='all'"><noscript><link rel="stylesheet" href="/css/syntax.css?v=3da135297e59"></noscript><title>Inkling - Thinking Machines Lab</title><meta property="og:title" content="Inkling"><meta property="og:description" content="Efficient open-weights models that reason over text, image, and audio. Fine-tune them on Tinker."><meta property="og:type" content="article"><meta property="og:url" content="https://thinkingmachines.ai/inkling/"><meta property="og:site_name" content="Thinking Machines Lab"><meta property="og:image" content="https://thinkingmachines.ai/inkling/images/cover-social-inkling-page.png"><meta property="og:image:width" content="2400"><meta property="og:image:height" content="1260"><meta property="og:image:alt" content="Inkling"><meta name="twitter:card" content="summary_large_image"><meta name="twitter:title" content="Inkling"><meta name="twitter:description" content="Efficient open-weights models that reason over text, image, and audio. Fine-tune them on Tinker."><meta name="twitter:image" content="https://thinkingmachines.ai/inkling/images/cover-social-inkling-page.png"><meta name="twitter:image:width" content="2400"><meta name="twitter:image:height" content="1260"><meta name="twitter:image:alt" content="Inkling"><link rel="preload" as="style" href="/css/model-card.min.59fba31ab10ca517bc342a339b22d1b8bab759f3f267691a3f54d1aadc765b14.css"><link rel="stylesheet" href="/css/model-card.min.59fba31ab10ca517bc342a339b22d1b8bab759f3f267691a3f54d1aadc765b14.css" media="print" onload="this.media='all'"><noscript><link rel="stylesheet" href="/css/model-card.min.59fba31ab10ca517bc342a339b22d1b8bab759f3f267691a3f54d1aadc765b14.css"></noscript><script src="/js/model-cards.min.f23924b4c54494b8a5434e2537e1be338e371f23564dcf7bdacb83d3035fbfc6.js" defer></script><link rel="preload" as="style" href="/css/inkling.min.7ab0098de2654c8f1e29575c93d6af1b035f4f65472a5a35b5e4afbec2245b3b.css"><link rel="stylesheet" href="/css/inkling.min.7ab0098de2654c8f1e29575c93d6af1b035f4f65472a5a35b5e4afbec2245b3b.css" media="print" onload="this.media='all'"><noscript><link rel="stylesheet" href="/css/inkling.min.7ab0098de2654c8f1e29575c93d6af1b035f4f65472a5a35b5e4afbec2245b3b.css"></noscript><script src="/js/inkling.min.a266f3bcd616178e7f350442f4ddfa0ee0aee4f663144f6244dbcdffed8d34e9.js" defer></script><link rel="stylesheet" href="/css/demo.min.3c363982bf09981d77eae40de4ade4ab8242cbf12b8eacec4bc0b029583d6efa.css"><script src="/js/demo-tabs.min.353e7322c19e6d4aded47c30958d8dd51fb6f3a8da2ed253be79286f3a4f9dc2.js" defer></script><noscript><style>.demo-panel[hidden]{display:block !important}.demo .demo-header{display:none}.demo .demo-body{height:auto;max-height:none}.demo .demo-body .demo-panel{max-height:none !important}</style></noscript><style>.post-group[data-astro-cid-jcsiju5y]{max-width:660px;padding-top:1.6rem;list-style-type:none;& li{padding-bottom:.4rem;position:relative}& li+li[data-astro-cid-jcsiju5y]{padding-top:.4rem}& li+li[data-astro-cid-jcsiju5y]:before{content:"";border-top:1px solid color-mix(in srgb, var(--fg4) 10%, transparent);pointer-events:none;position:absolute;top:0;left:0;right:0}& .post-item-link{color:inherit;cursor:pointer;border-radius:8px;margin:0 -1.6rem;padding:1rem 1.6rem;line-height:1.4;text-decoration:none;display:block;@media (width>=768px){grid-template-columns:7rem 1fr;align-items:baseline;gap:1.5rem;margin:0 -1.5rem;padding:1.2rem 1.5rem;display:grid}}& .post-info{flex-direction:column;align-items:flex-start;gap:.25rem;display:flex;@media (width>=768px){grid-area:1/2}}& .post-title{color:var(--fg);text-wrap:balance;text-underline-offset:2px;font-size:17px;font-weight:500;text-decoration:underline #0000;transition:color .2s,text-decoration-color .2s;display:inline-block}& .post-subtitle{color:var(--fg1);font-size:.95em;line-height:1.4;transition:color .2s}& .author-date{color:var(--fg4);font-size:15px;font-family:var(--font-sans);display:block}& time{font-size:15px;font-family:var(--font-sans);color:var(--fg4);margin-top:.25rem;transition:color .2s;display:block;@media (width>=768px){text-align:left;white-space:nowrap;grid-area:1/1;margin-top:0}}@media (hover:hover){& .post-item-link:hover{& .post-title{color:var(--fg);text-decoration-color:color-mix(in srgb, var(--fg2) 20%, transparent)}& .post-subtitle{color:var(--fg)}& .author-date,& time{color:var(--fg3)}}}}
svg[data-astro-cid-t4uphou2]{cursor:none}@media (width<=640px){svg[data-astro-cid-t4uphou2]{aspect-ratio:4/3}}
.btn[data-astro-cid-ekguhzzh]{appearance:none;min-height:40px;color:var(--fg);font-family:var(--font-sans);cursor:pointer;transition:transform .45s cubic-bezier(.34,1.56,.64,1);transition-timing-function:linear(0, .44 10%, .94 20%, 1.32 30%, 1.5 40%, 1.28 50%, .94 60%, .75 70%, .82 78%, .96 86%, 1.03 93%, 1);border:none;border-radius:100px;justify-content:center;align-items:center;gap:.5rem;padding:0 18px;font-size:15px;font-weight:400;line-height:1;text-decoration:none;display:inline-flex;position:relative;overflow:hidden}.btn[data-astro-cid-ekguhzzh]:before{content:"";opacity:0;pointer-events:none;border-radius:inherit;background:#000;transition:opacity .18s;position:absolute;inset:0}.btn[data-astro-cid-ekguhzzh]:active{transition:transform .12s cubic-bezier(.4,0,.2,1);transform:scale(.96)}.btn[data-astro-cid-ekguhzzh][disabled]{opacity:.6;cursor:not-allowed}.btn[data-astro-cid-ekguhzzh].primary{color:#000;background:linear-gradient(#f3f2f2 0%,#e6e6e6 87.5%)}.btn[data-astro-cid-ekguhzzh].primary:hover:before{opacity:.04}.btn[data-astro-cid-ekguhzzh].primary:active:before{opacity:.08}.btn[data-astro-cid-ekguhzzh].plain{color:var(--fg);background:0 0}.btn[data-astro-cid-ekguhzzh].plain:before{display:none}.btn[data-astro-cid-ekguhzzh].plain:hover,.btn[data-astro-cid-ekguhzzh].plain:focus-visible{background:#0000000d}.btn[data-astro-cid-ekguhzzh].plain:active{background:#00000014}.btn-with-chevron[data-astro-cid-ekguhzzh]{gap:.25rem}.btn-with-chevron[data-astro-cid-ekguhzzh] svg[data-astro-cid-ekguhzzh]{width:16px;height:16px;transition:transform .2s ease-in-out}.btn-with-chevron[data-astro-cid-ekguhzzh] svg[data-astro-cid-ekguhzzh] path[data-astro-cid-ekguhzzh]{opacity:.4;transition:opacity .2s ease-in-out}.btn-with-chevron[data-astro-cid-ekguhzzh]:hover svg[data-astro-cid-ekguhzzh]{transform:translate(3px)}.btn-with-chevron[data-astro-cid-ekguhzzh]:hover svg[data-astro-cid-ekguhzzh] path[data-astro-cid-ekguhzzh]{opacity:.7}@media (prefers-reduced-motion:reduce){.btn[data-astro-cid-ekguhzzh],.btn[data-astro-cid-ekguhzzh]:active,.btn-with-chevron[data-astro-cid-ekguhzzh] svg[data-astro-cid-ekguhzzh]{transition:none}}
</style></head><body><div id="wrapper"><header id="header"><!-- Navbar for all pages --><div id="site-header" class="site-header"><div class="navbar-logo"><a href="/">THINKING MACHINES</a></div><!-- Navigation group on the right --><div class="nav-group"><nav class="top-nav"><a href="/inkling/">Inkling</a><a href="/tinker/">Tinker</a><a href="/blog/">Connectionism</a><a href="/news/">News</a><a href="/#join-us">Join us</a></nav><button id="menu-toggle" class="hamburger" aria-label="Open menu" aria-expanded="false" aria-controls="mobile-menu" type="button"><span></span><span></span><span></span></button></div></div><!-- Mobile overlay menu --><div id="mobile-menu" class="mobile-menu-overlay" aria-hidden="true" hidden><nav class="mobile-menu-nav" role="navigation" aria-label="Mobile"><a href="/">Home</a><a href="/inkling/">Inkling</a><a href="/tinker/">Tinker</a><a href="/blog/">Connectionism</a><a href="/news/">News</a><a href="/#join-us">Join us</a></nav></div></header><main id="main" class="content post-inkling"><article class="content"><div class="inkling-hero-section"><section class="inkling-hero"><div class="inkling inkling-artwork inkling--deferred" data-recipe="{&quot;seeds&quot;:[4080338587,2713983725,38430317,1770525004,3398537396,930963614,338165897,2228226371,1966101379,3760791918,4151914198],&quot;grid&quot;:24,&quot;count&quot;:3,&quot;mode&quot;:&quot;color&quot;,&quot;ixStyle&quot;:&quot;normal&quot;,&quot;round&quot;:0.8,&quot;declutter&quot;:2,&quot;gridlines&quot;:false}" data-interactive="true"></div><p class="inkling-eyebrow">Inkling</p><h1 class="inkling-title">Open weights, ready to tinker</h1><p class="inkling-subtitle">Efficient open models for text, images, and audio, available to fine-tune.</p><div class="inkling-cta-row"><a class="btn primary" href="https://tinker.thinkingmachines.ai/playground?utm_source=blog&amp;utm_campaign=inkling_model_release" target="_blank" rel="noopener" data-astro-cid-ekguhzzh="true">Try on Tinker</a><a class="btn plain btn-with-chevron" href="https://huggingface.co/collections/thinkingmachines/inkling" target="_blank" rel="noopener" data-astro-cid-ekguhzzh="true">Hugging Face<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" data-astro-cid-ekguhzzh><path fill-rule="evenodd" clip-rule="evenodd" d="M6.56701 3.34347C6.78189 3.14201 7.1194 3.1529 7.32086 3.36779L11.3208 7.63446C11.5131 7.8396 11.5131 8.15884 11.3208 8.36399L7.32086 12.6306C7.1194 12.8456 6.78189 12.8565 6.56701 12.655C6.35212 12.4535 6.34123 12.116 6.54269 11.9011L10.2007 7.99922L6.54269 4.09733C6.34123 3.88244 6.35212 3.54492 6.56701 3.34347Z" fill="currentColor" data-astro-cid-ekguhzzh></path></svg></a></div></section></div>
<section class="inkling-section inkling-models-section"><h2 class="inkling-h2">Two sizes of Inkling</h2><div class="model-cards"><article class="model-card model-card--inkling" style="--waffle-accent: #0155BF"><div class="model-card-head"><h3 class="model-card-title"><a class="model-card-titlelink" href="/news/introducing-inkling/">Inkling</a></h3></div><p class="model-card-body">Our most capable open model, with native reasoning over text, images, and audio.</p><div class="model-card-waffle"><div class="model-card-waffle-grid" style="--waffle-cols: 40" role="img" data-waffle-sequence="0" aria-label="Inkling: 975 billion total parameters, drawn as 200 of 200 blocks, of which 8 represent the 41 billion activated per token"><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span></div></div><dl class="model-card-specs"><dt class="model-card-spec-label">Parameters</dt><dd class="model-card-spec-value model-card-spec-value--keyed"><span class="model-card-key"><span class="model-card-key-chip model-card-key-chip--total"></span>975B total</span> <span class="model-card-key"><span class="model-card-key-chip model-card-key-chip--active"></span>41B active</span></dd><dt class="model-card-spec-label">Architecture</dt><dd class="model-card-spec-value">Mixture of Experts</dd><dt class="model-card-spec-label">Context</dt><dd class="model-card-spec-value">1M tokens (64K / 256K on Tinker)</dd><dt class="model-card-spec-label">Input</dt><dd class="model-card-spec-value">Text, Image, Audio</dd></dl><div class="model-card-links"><a class="btn primary model-card-cta" href="/news/introducing-inkling/" aria-label="Learn more about Inkling" data-astro-cid-ekguhzzh="true">Learn more</a> <a class="btn plain btn-with-chevron" href="/model-card/inkling/" data-astro-cid-ekguhzzh="true">Model card<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" data-astro-cid-ekguhzzh><path fill-rule="evenodd" clip-rule="evenodd" d="M6.56701 3.34347C6.78189 3.14201 7.1194 3.1529 7.32086 3.36779L11.3208 7.63446C11.5131 7.8396 11.5131 8.15884 11.3208 8.36399L7.32086 12.6306C7.1194 12.8456 6.78189 12.8565 6.56701 12.655C6.35212 12.4535 6.34123 12.116 6.54269 11.9011L10.2007 7.99922L6.54269 4.09733C6.34123 3.88244 6.35212 3.54492 6.56701 3.34347Z" fill="currentColor" data-astro-cid-ekguhzzh></path></svg></a></div></article><article class="model-card model-card--inklingsmall" style="--waffle-accent: #8F3F71"><div class="model-card-head"><h3 class="model-card-title"><a class="model-card-titlelink" href="/news/inkling-small/">Inkling-Small</a></h3><span class="model-card-badge">NEW</span></div><p class="model-card-body">Matches its larger sibling on many benchmarks at a quarter of the size, with lower cost and latency.</p><div class="model-card-waffle"><div class="model-card-waffle-grid" style="--waffle-cols: 40" role="img" data-waffle-sequence="1" aria-label="Inkling-Small: 276 billion total parameters, drawn as 55 of 200 blocks, of which 2 represent the 12 billion activated per token"><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-active"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-filled"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span><span class="model-card-waffle-cell is-empty"></span></div></div><dl class="model-card-specs"><dt class="model-card-spec-label">Parameters</dt><dd class="model-card-spec-value model-card-spec-value--keyed"><span class="model-card-key"><span class="model-card-key-chip model-card-key-chip--total"></span>276B total</span> <span class="model-card-key"><span class="model-card-key-chip model-card-key-chip--active"></span>12B active</span></dd><dt class="model-card-spec-label">Architecture</dt><dd class="model-card-spec-value">Mixture of Experts</dd><dt class="model-card-spec-label">Context</dt><dd class="model-card-spec-value">1M tokens (64K / 256K on Tinker)</dd><dt class="model-card-spec-label">Input</dt><dd class="model-card-spec-value">Text, Image, Audio</dd></dl><div class="model-card-links"><a class="btn primary model-card-cta" href="/news/inkling-small/" aria-label="Learn more about Inkling-Small" data-astro-cid-ekguhzzh="true">Learn more</a> <a class="btn plain btn-with-chevron" href="/model-card/inkling-small/" data-astro-cid-ekguhzzh="true">Model card<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" data-astro-cid-ekguhzzh><path fill-rule="evenodd" clip-rule="evenodd" d="M6.56701 3.34347C6.78189 3.14201 7.1194 3.1529 7.32086 3.36779L11.3208 7.63446C11.5131 7.8396 11.5131 8.15884 11.3208 8.36399L7.32086 12.6306C7.1194 12.8456 6.78189 12.8565 6.56701 12.655C6.35212 12.4535 6.34123 12.116 6.54269 11.9011L10.2007 7.99922L6.54269 4.09733C6.34123 3.88244 6.35212 3.54492 6.56701 3.34347Z" fill="currentColor" data-astro-cid-ekguhzzh></path></svg></a></div></article></div></section>
<section class="inkling-section"><h2 class="inkling-h2 inkling-h2--benchmark">Well-rounded, generalist models</h2><div class="spider-switch" data-demo><div class="demo-tabs spider-switch-tabs" role="tablist" aria-label="Choose a model to compare"><button class="demo-tab" type="button" role="tab" data-demo-tab id="spider-switch-79d4c5f4-tab-inkling" aria-controls="spider-switch-79d4c5f4-panel-inkling" aria-selected="true" tabindex="0">Inkling</button><button class="demo-tab" type="button" role="tab" data-demo-tab id="spider-switch-79d4c5f4-tab-small" aria-controls="spider-switch-79d4c5f4-panel-small" aria-selected="false" tabindex="-1">Inkling-Small</button></div><div class="demo-panel spider-switch-panel" id="spider-switch-79d4c5f4-panel-inkling" role="tabpanel" aria-labelledby="spider-switch-79d4c5f4-tab-inkling"><style>  .benchmark-radar {
    --radar-inkling: #0155BF;

    --radar-nemotron: #EF402C;
    --radar-glm: #21B3FF;
    --radar-gpt: #F7A224;
    --radar-claude: #0E9972;

    --radar-grid: color-mix(in srgb, var(--fg4) 17%, var(--bg) 83%);
    --radar-axis: color-mix(in srgb, var(--fg4) 25%, var(--bg) 75%);
    color: var(--fg);
    font-family: var(--font-sans);
    margin: 0 auto;
    max-width: 860px;
    position: relative;
    width: 100%;
  }

  @media (min-width: 1171px) {
    #main article.content .benchmark-radar-figure {
        margin-left: calc((100% - 860px) / 2);
      margin-right: 0;
      max-width: none;
        width: 860px;
    }
  }

  #main article.content .benchmark-radar-legend {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    font-size: 15px;
    gap: 0.35rem 0.3rem;
    justify-content: center;
    margin: 0.7rem auto 0;
    position: relative;
  }

  #main article.content .benchmark-radar-legend-pill {
    background: color-mix(in srgb, var(--fg) 10%, var(--bg) 90%);
    border-radius: 999px;
    height: 0;
    left: 0;
    opacity: 0;
    pointer-events: none;
    position: absolute;
    top: 0;
    transform: translate3d(0, 0, 0);
    transition:
      opacity 120ms ease,
      transform 280ms cubic-bezier(0.22, 1, 0.36, 1),
      width 240ms cubic-bezier(0.22, 1, 0.36, 1),
      height 240ms cubic-bezier(0.22, 1, 0.36, 1);
    width: 0;
    will-change: transform, width, height;
    z-index: 0;
  }

  #main article.content .benchmark-radar-legend-pill.is-visible {
    opacity: 1;
  }

  #main article.content .benchmark-radar-legend-item {
    -webkit-appearance: none;
    align-items: center;
    appearance: none;
    background: transparent;
    border: 0;
    border-radius: 999px;
    color: var(--fg);
    cursor: pointer;
    display: inline-flex;
    font: inherit;
    font-weight: 400;
    gap: 0.28rem;
    line-height: 1;
    margin: 0;
    padding: 0.62rem 0.9rem;
    position: relative;
    white-space: nowrap;
    z-index: 1;
  }

  html.js #main article.content .benchmark-radar-legend-item:hover,
  html.js #main article.content .benchmark-radar-legend-item:focus-visible {
    background: transparent;
  }

  #main article.content .benchmark-radar-legend-item:focus-visible {
    outline: 2px solid color-mix(in srgb, var(--fg) 42%, transparent 58%);
    outline-offset: 2px;
  }

  html:not(.js) #main article.content .benchmark-radar-legend-item:hover,
  html:not(.js) #main article.content .benchmark-radar-legend-item:focus-visible {
    background: color-mix(in srgb, var(--fg4) 11%, var(--bg) 89%);
  }

  .benchmark-radar-key {
    border-radius: 50%;
    box-sizing: border-box;
    display: inline-block;
    flex: 0 0 auto;
    height: 9px;
    width: 9px;
  }

  /* Every key is a hollow ring in its model's colour; the hero's is filled in.
     Keyed on --hero rather than on a slug so the landing page's two charts,
     whose styles share one document, cannot fill each other's keys. */
  .benchmark-radar-key {
    background: transparent;
    border: 2px solid transparent;
  }
  .benchmark-radar-key.benchmark-radar-key--inkling { border-color: var(--radar-inkling); }
  .benchmark-radar-key.benchmark-radar-key--inklingsmall { border-color: var(--radar-inklingsmall); }
  .benchmark-radar-key--hero.benchmark-radar-key--inkling {
    background: var(--radar-inkling);
    border-color: transparent;
  }
  .benchmark-radar-key--hero.benchmark-radar-key--inklingsmall {
    background: var(--radar-inklingsmall);
    border-color: transparent;
  }
  .benchmark-radar-key.benchmark-radar-key--nemotron { border-color: var(--radar-nemotron); }
  .benchmark-radar-key.benchmark-radar-key--glm { border-color: var(--radar-glm); }
  .benchmark-radar-key.benchmark-radar-key--gpt { border-color: var(--radar-gpt); }
  .benchmark-radar-key.benchmark-radar-key--claude { border-color: var(--radar-claude); }



  .benchmark-radar-scroll {
    display: block;
    max-width: 100%;
    min-width: 0;
    overflow-x: auto;
    overflow-y: hidden;
    overscroll-behavior-x: contain;
    width: 100%;
    -webkit-overflow-scrolling: touch;
  }

  .benchmark-radar-scroll::-webkit-scrollbar { height: 6px; }
  .benchmark-radar-scroll::-webkit-scrollbar-thumb {
    background: color-mix(in srgb, var(--fg4) 45%, var(--bg) 55%);
    border-radius: 3px;
  }

  .benchmark-radar-svg {
    min-width: 720px;
    overflow: visible;
  }

  .benchmark-radar-svg text {
    fill: var(--fg2);
    font-family: var(--font-sans);
    font-size: 13px;
  }

  .benchmark-radar-svg .benchmark-radar-tick {
    fill: var(--fg4);
    font-size: 13px;
  }

  .benchmark-radar-svg .benchmark-radar-tick--detail {
    opacity: 0;
    transition: opacity 160ms ease;
  }

  .benchmark-radar.is-radar-filtered .benchmark-radar-tick--detail {
    opacity: 1;
  }

  .benchmark-radar-label {
    pointer-events: none;
    transform-box: fill-box;
    transform-origin: center;
    transition: fill 140ms ease, transform 180ms cubic-bezier(0.22, 1, 0.36, 1);
  }

  .benchmark-radar-label-title {
    fill: var(--fg);
  }

  .benchmark-radar-label-category {
    fill: var(--fg4);
    opacity: 0.82;
  }

  .benchmark-radar-label-target {
    cursor: default;
  }

  .benchmark-radar-label-hit {
    fill: transparent;
    pointer-events: all;
  }

  .benchmark-radar-label-target.is-hovered .benchmark-radar-label {
    fill: var(--fg);
    transform: scale(1.16);
  }

  .benchmark-radar-grid {
    fill: none;
    stroke: var(--radar-grid);
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
  }

  .benchmark-radar-axis {
    stroke: var(--radar-axis);
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
  }

  .benchmark-radar-area { pointer-events: none; }
  .benchmark-radar-segment,
  .benchmark-radar-point {
    vector-effect: non-scaling-stroke;
  }

  .benchmark-radar-segment {
    fill: none;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-width: 2;
  }

  .benchmark-radar-series--inkling .benchmark-radar-segment { stroke-width: 2.8; }
  .benchmark-radar-point {
    fill: var(--bg);
    stroke-width: 2.2;
  }

  .benchmark-radar-series--hero.benchmark-radar-series--inkling .benchmark-radar-point {
    fill: var(--radar-inkling);
    stroke: var(--bg) !important;
    stroke-width: 1.5;
  }

  .benchmark-radar-series {
    transition: opacity 180ms ease;
    transform-box: view-box;
    transform-origin: 50% 50%;
  }

  .benchmark-radar-area {
    transition: fill-opacity 180ms ease, opacity 180ms ease;
  }

  .benchmark-radar.is-radar-filtered .benchmark-radar-series {
    opacity: 0.18 !important;
  }

  .benchmark-radar.is-radar-filtered[data-active-model="inkling"] .benchmark-radar-series--inkling,
  .benchmark-radar.is-radar-filtered[data-active-model="nemotron"] .benchmark-radar-series--nemotron,
  .benchmark-radar.is-radar-filtered[data-active-model="glm"] .benchmark-radar-series--glm,
  .benchmark-radar.is-radar-filtered[data-active-model="gpt"] .benchmark-radar-series--gpt,
  .benchmark-radar.is-radar-filtered[data-active-model="claude"] .benchmark-radar-series--claude {
    opacity: 1 !important;
  }


  .benchmark-radar.is-radar-filtered .benchmark-radar-area {
    opacity: 0 !important;
  }

  .benchmark-radar.is-radar-filtered[data-active-model="inkling"] .benchmark-radar-series--inkling .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="nemotron"] .benchmark-radar-series--nemotron .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="glm"] .benchmark-radar-series--glm .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="gpt"] .benchmark-radar-series--gpt .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="claude"] .benchmark-radar-series--claude .benchmark-radar-area {
    fill-opacity: 0.2;
    opacity: 1 !important;
  }


  .benchmark-radar.is-radar-filtered[data-active-model="inkling"] .benchmark-radar-series--inkling .benchmark-radar-area {
    fill-opacity: 0.26;
  }

  #main article.content .benchmark-radar-tooltip {
    background: #111;
    border: 0;
    border-radius: 4px;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.16);
    color: #fff;
    font-family: var(--font-sans);
    font-size: 11px;
    font-weight: 400;
    left: 0;
    line-height: 1;
    max-width: calc(100% - 16px);
    min-width: 190px;
    opacity: 0;
    padding: 0.45rem 0.55rem;
    pointer-events: none;
    position: absolute;
    top: 0;
    transform: translateY(4px);
    transition: opacity 140ms ease, transform 140ms ease;
    width: 210px;
    z-index: 3;
  }

  #main article.content .benchmark-radar-tooltip[aria-hidden="false"] {
    opacity: 1;
    transform: translateY(0);
  }

  #main article.content .benchmark-radar-tooltip-title {
    color: #fff;
    font-weight: 500;
    margin-bottom: 0.4rem;
  }

  #main article.content .benchmark-radar-tooltip [data-radar-tooltip-rows] {
    display: grid;
    gap: 0.15rem;
    margin: 0;
  }

  #main article.content .benchmark-radar-tooltip-row {
    align-items: center;
    display: grid;
    gap: 0.4rem;
    grid-template-columns: 8px minmax(0, 1fr) auto;
    line-height: 1.35;
    margin: 0;
  }

  #main article.content .benchmark-radar-tooltip-dot {
    border-radius: 50%;
    height: 8px;
    width: 8px;
  }

  #main article.content .benchmark-radar-tooltip-value {
    font-variant-numeric: tabular-nums;
    text-align: right;
  }

  #main article.content .benchmark-radar-tooltip-value.is-missing {
    color: rgba(255, 255, 255, 0.62);
  }

  .benchmark-radar-series--hero .benchmark-radar-point {
    transform-box: fill-box;
    transform-origin: center;
  }

  html.js .benchmark-radar.is-radar-pending .benchmark-radar-series--hero .benchmark-radar-point {
    opacity: 0;
    transform: scale(0);
  }

  html.js .benchmark-radar.is-radar-pending .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment) {
    opacity: 0;
  }

  html.js .benchmark-radar.is-radar-pending .benchmark-radar-series:not(.benchmark-radar-series--hero) {
    opacity: 0;
    transform: scale(0.99);
  }

  html.js .benchmark-radar.is-radar-visible .benchmark-radar-series--hero .benchmark-radar-point {
    animation: benchmark-radar-point-in 260ms cubic-bezier(0.2, 0.78, 0.22, 1) both;
    animation-delay: var(--radar-point-delay, 0ms);
  }

  html.js .benchmark-radar.is-radar-visible .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment) {
    animation: benchmark-radar-hero-shape-in 320ms ease-out 820ms both;
  }

  html.js .benchmark-radar.is-radar-visible .benchmark-radar-series:not(.benchmark-radar-series--hero) {
    animation: benchmark-radar-comparison-in 480ms ease-out 820ms both;
  }

  @keyframes benchmark-radar-point-in {
    0% { opacity: 0; transform: scale(0); }
    70% { opacity: 1; transform: scale(1.2); }
    100% { opacity: 1; transform: scale(1); }
  }

  @keyframes benchmark-radar-hero-shape-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes benchmark-radar-comparison-in {
    from { opacity: 0; transform: scale(0.99); }
    to { opacity: 1; transform: scale(1); }
  }

  @media (prefers-reduced-motion: reduce) {
    #main article.content .benchmark-radar-legend-pill {
      transition: none;
    }

    .benchmark-radar-label {
      transition: none;
    }

    html.js .benchmark-radar.is-radar-pending .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment, .benchmark-radar-point),
    html.js .benchmark-radar.is-radar-visible .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment, .benchmark-radar-point),
    html.js .benchmark-radar.is-radar-pending .benchmark-radar-series:not(.benchmark-radar-series--hero),
    html.js .benchmark-radar.is-radar-visible .benchmark-radar-series:not(.benchmark-radar-series--hero) {
      animation: none;
      opacity: 1;
      transform: none;
    }
  }

  @media (max-width: 620px) {
    #main article.content .benchmark-radar-legend {
      justify-content: center;
    }

    /* Phones: scale the whole chart to fit instead of horizontal-scrolling a
       radar (you can't read a shape you can only see half of). Labels are
       enlarged in user units to stay legible once the SVG is scaled down. */
    .benchmark-radar-scroll {
      overflow-x: visible;
    }

    .benchmark-radar-svg {
      min-width: 0;
      width: 100%;
      height: auto;
    }

    .benchmark-radar-svg text {
      font-size: 21px;
    }

    .benchmark-radar-svg .benchmark-radar-tick {
      font-size: 19px;
    }
  }</style><div class="benchmark-radar-figure"><div class="benchmark-radar" id="benchmark-radar-inkling"><div class="benchmark-radar-scroll"><svg class="benchmark-radar-svg" data-benchmark-radar viewBox="0 0 860 600" role="img" aria-label="Inkling benchmark comparison" aria-describedby="benchmark-radar-desc-inkling"><desc id="benchmark-radar-desc-inkling">Spider chart comparing Inkling, Nemotron 3 Ultra, GLM 5.2, GPT 5.6 Sol, and Claude Fable 5 on ten evaluations scored from zero to one hundred. Inkling is shown with the heavier cobalt line. Evaluations without a reported model score are plotted at zero. Hover an evaluation to compare every model&#39;s score.</desc><g data-radar-grid></g><g data-radar-axes></g><g data-radar-series></g><g data-radar-labels></g></svg></div><div class="benchmark-radar-legend" aria-label="Model legend"><span class="benchmark-radar-legend-pill" data-radar-legend-pill aria-hidden="true"></span><button class="benchmark-radar-legend-item" type="button" data-radar-model="inkling"><span class="benchmark-radar-key benchmark-radar-key--inkling benchmark-radar-key--hero" aria-hidden="true"></span>Inkling</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="nemotron"><span class="benchmark-radar-key benchmark-radar-key--nemotron" aria-hidden="true"></span>Nemotron 3 Ultra</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="glm"><span class="benchmark-radar-key benchmark-radar-key--glm" aria-hidden="true"></span>GLM 5.2</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="gpt"><span class="benchmark-radar-key benchmark-radar-key--gpt" aria-hidden="true"></span>GPT 5.6 Sol</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="claude"><span class="benchmark-radar-key benchmark-radar-key--claude" aria-hidden="true"></span>Claude Fable 5</button></div><div class="benchmark-radar-tooltip tooltip" data-radar-tooltip role="tooltip" aria-hidden="true"><div class="benchmark-radar-tooltip-title" data-radar-tooltip-title></div><div data-radar-tooltip-rows></div></div></div></div><script>(function(){const variant = "inkling";
const hero = "inkling";
const metrics = [{"name":"HLE (no tools)","category":"Reasoning"},{"name":"AIME 2026","category":"Reasoning"},{"name":"SWE-bench Pro (Public)","category":"Agentic (coding)"},{"name":"Terminal Bench 2.1","category":"Agentic (coding)"},{"name":"MCP Atlas","category":"Agentic (general)"},{"name":"SimpleQA Verified","category":"Factuality"},{"name":"AA Omniscience (acc)","category":"Factuality"},{"name":"AudioMC","category":"Audio"},{"name":"MMMU Pro","category":"Vision"},{"name":"IFBench","category":"Chat"}];
const models = [{"name":"Inkling","slug":"inkling","color":"var(--radar-inkling)","dash":"","fillOpacity":0.12,"values":[29.7,97.1,54.3,63.8,76,43.9,40,56.6,73.5,79.8]},{"name":"Nemotron 3 Ultra","slug":"nemotron","color":"var(--radar-nemotron)","dash":"7 4","fillOpacity":0.035,"values":[26.6,94.2,46.4,56.4,42.7,32.4,22,null,null,81.4]},{"name":"GLM 5.2","slug":"glm","color":"var(--radar-glm)","dash":"2 4","fillOpacity":0.035,"values":[40.1,99.2,62.1,82.7,77.8,38.1,25,null,null,73.3]},{"name":"GPT 5.6 Sol","slug":"gpt","color":"var(--radar-gpt)","dash":"9 4 2 4","fillOpacity":0.035,"values":[47.2,99.9,64.6,89.5,81.8,71.6,58,null,83,72.7]},{"name":"Claude Fable 5","slug":"claude","color":"var(--radar-claude)","dash":"6 4","fillOpacity":0.035,"values":[53.3,99.9,80,84.6,83.3,68.3,61,null,84.2,63.5]}];

  (() => {
    const radar = document.getElementById('benchmark-radar-' + variant);
    const heroSlug = hero;
    const svg = radar && radar.querySelector('[data-benchmark-radar]');
    if (!radar || !svg) return;

    const initialize = () => {
      if (svg.dataset.ready === 'true') return;
      svg.dataset.ready = 'true';

    const ns = 'http://www.w3.org/2000/svg';
    const center = { x: 430, y: 300 };
    const radius = 220;

    const create = (tag, attrs = {}) => {
      const node = document.createElementNS(ns, tag);
      Object.entries(attrs).forEach(([key, value]) => node.setAttribute(key, value));
      return node;
    };
    const point = (index, value, extra = 0) => {
      const angle = -Math.PI / 2 + index * Math.PI * 2 / metrics.length;
      const distance = radius * value / 100 + extra;
      return {
        x: center.x + Math.cos(angle) * distance,
        y: center.y + Math.sin(angle) * distance,
        ux: Math.cos(angle),
        uy: Math.sin(angle)
      };
    };
    const modelGeometry = models.map((model) => {
      const points = model.values.map((value, index) => {
        const score = value ?? 0;
        return { index, value: score, ...point(index, score) };
      });
      const segments = points.map((from, pointIndex) => ({
        from,
        to: points[(pointIndex + 1) % points.length]
      }));
      return { model, points, segments };
    });
    const pointsAttr = (value) => metrics.map((_, index) => {
      const p = point(index, value);
      return `${p.x.toFixed(2)},${p.y.toFixed(2)}`;
    }).join(' ');

    const grid = svg.querySelector('[data-radar-grid]');
    [20, 40, 60, 80, 100].forEach((value) => {
      grid.append(create('polygon', { class: 'benchmark-radar-grid', points: pointsAttr(value) }));
      const p = point(0, value);
      const tick = create('text', {
        class: 'benchmark-radar-tick benchmark-radar-tick--detail',
        x: p.x + 7,
        y: p.y + 16
      });
      tick.textContent = value;
      grid.append(tick);
    });

    const axes = svg.querySelector('[data-radar-axes]');
    metrics.forEach((_, index) => {
      const p = point(index, 100);
      axes.append(create('line', {
        class: 'benchmark-radar-axis',
        x1: center.x, y1: center.y, x2: p.x, y2: p.y
      }));
    });

    const seriesLayer = svg.querySelector('[data-radar-series]');
    [...modelGeometry].reverse().forEach(({ model, points }) => {
      const group = create('g', {
        class: `benchmark-radar-series benchmark-radar-series--${model.slug}` + (model.slug === heroSlug ? ' benchmark-radar-series--hero' : '')
      });

      const areaPath = points.map((p, index) => `${index === 0 ? 'M' : 'L'} ${p.x.toFixed(2)} ${p.y.toFixed(2)}`).join(' ') + ' Z';
      group.append(create('path', {
        class: 'benchmark-radar-area',
        d: areaPath,
        fill: model.color,
        'fill-opacity': model.fillOpacity
      }));

      points.forEach((from, pointIndex) => {
        const to = points[(pointIndex + 1) % points.length];
        const segment = create('line', {
          class: 'benchmark-radar-segment',
          x1: from.x, y1: from.y, x2: to.x, y2: to.y,
          stroke: model.color
        });
        if (model.dash) segment.setAttribute('stroke-dasharray', model.dash);
        group.append(segment);
      });

      points.forEach((p) => {
        const marker = create('circle', {
          class: 'benchmark-radar-point',
          cx: p.x, cy: p.y, r: model.slug === heroSlug ? 5 : 4.5,
          stroke: model.color,
          style: `--radar-point-delay: ${p.index * 55}ms`
        });
        group.append(marker);
      });
      seriesLayer.append(group);
    });

    const labels = svg.querySelector('[data-radar-labels]');
    metrics.forEach((metric, index) => {
      const p = point(index, 100, 28);
      const target = create('g', {
        class: 'benchmark-radar-label-target',
        'data-radar-metric-index': index
      });
      const label = create('text', { class: 'benchmark-radar-label', x: p.x, y: p.y + 5 });
      label.setAttribute('text-anchor', p.ux > 0.24 ? 'start' : p.ux < -0.24 ? 'end' : 'middle');
      const title = create('tspan', { class: 'benchmark-radar-label-title', x: p.x, dy: 0 });
      title.textContent = metric.name;
      const category = create('tspan', { class: 'benchmark-radar-label-category', x: p.x, dy: '1.3em' });
      category.textContent = metric.category;
      label.append(title, category);
      target.append(label);
      labels.append(target);

      /* A long name on a side axis can run past the SVG box once the phone
         breakpoint enlarges label text in user units — "SWE-bench Pro (Public)"
         overhung the right edge by ~30 units at 390px wide. Only when a label
         actually overflows, break its trailing parenthetical onto a second line.
         Measuring rather than hard-coding the one name keeps the qualifier and
         the enlarged font, leaves every label that already fits alone, and
         re-evaluates itself if names, font sizes or breakpoints change later.
         Note Terminal Bench 2.1 shares this side with only ~20 units to spare,
         so shifting the labels outward is not an option. */
      const box = svg.viewBox.baseVal;
      const first = label.getBBox();
      const parenthetical = metric.name.match(/^(.*\S)\s+(\(.*\))$/);
      if (parenthetical && (first.x < box.x || first.x + first.width > box.x + box.width)) {
        title.textContent = parenthetical[1];
        const rest = create('tspan', { class: 'benchmark-radar-label-title', x: p.x, dy: '1.15em' });
        rest.textContent = parenthetical[2];
        label.insertBefore(rest, category);
      }

      const bounds = label.getBBox();
      target.insertBefore(create('rect', {
        class: 'benchmark-radar-label-hit',
        x: bounds.x - 8,
        y: bounds.y - 6,
        width: bounds.width + 16,
        height: bounds.height + 12,
        rx: 4
      }), label);
    });

    const tooltip = radar.querySelector('[data-radar-tooltip]');
    const tooltipTitle = tooltip.querySelector('[data-radar-tooltip-title]');
    const tooltipRows = tooltip.querySelector('[data-radar-tooltip-rows]');
    const scrollContainer = radar.querySelector('.benchmark-radar-scroll');
    let activeMetricIndex = -1;

    const formatTooltipValue = (metric, value) => {
      if (value != null) return value.toFixed(1);
      if (metric.category === 'Audio') return 'audio not supported';
      if (metric.category === 'Vision') return 'vision not supported';
      return '0.0 - not reported';
    };

    models.forEach((model) => {
      const row = document.createElement('div');
      row.className = 'benchmark-radar-tooltip-row';
      row.dataset.radarTooltipModel = model.slug;

      const dot = document.createElement('span');
      dot.className = 'benchmark-radar-tooltip-dot';
      dot.style.background = model.color;
      dot.setAttribute('aria-hidden', 'true');

      const name = document.createElement('span');
      name.textContent = model.name;

      const value = document.createElement('span');
      value.className = 'benchmark-radar-tooltip-value';
      value.dataset.radarTooltipValue = '';

      row.append(dot, name, value);
      tooltipRows.append(row);
    });

    const positionTooltip = (metricIndex) => {
      if (metricIndex < 0) return;
      const rootRect = radar.getBoundingClientRect();
      const svgRect = svg.getBoundingClientRect();
      const viewBox = svg.viewBox.baseVal;
      const anchor = point(metricIndex, 100, 24);
      const anchorX = svgRect.left - rootRect.left + (anchor.x - viewBox.x) * svgRect.width / viewBox.width;
      const anchorY = svgRect.top - rootRect.top + (anchor.y - viewBox.y) * svgRect.height / viewBox.height;
      const gap = 12;
      const inset = 8;
      const tooltipWidth = tooltip.offsetWidth;
      const tooltipHeight = tooltip.offsetHeight;
      const rootWidth = radar.clientWidth;
      const rootHeight = radar.clientHeight;
      const placeBelow = anchorY < rootHeight * 0.48;
      const left = Math.min(Math.max(anchorX - tooltipWidth / 2, inset), rootWidth - tooltipWidth - inset);
      const preferredTop = placeBelow ? anchorY + gap : anchorY - tooltipHeight - gap;
      const top = Math.min(Math.max(preferredTop, inset), rootHeight - tooltipHeight - inset);
      tooltip.style.left = `${left}px`;
      tooltip.style.top = `${top}px`;
    };

    const showMetricTooltip = (metricIndex) => {
      activeMetricIndex = metricIndex;
      tooltipTitle.textContent = metrics[metricIndex].name;
      models.forEach((model) => {
        const value = model.values[metricIndex];
        const formattedValue = formatTooltipValue(metrics[metricIndex], value);
        const valueElement = tooltip.querySelector(`[data-radar-tooltip-model="${model.slug}"] [data-radar-tooltip-value]`);
        valueElement.textContent = formattedValue;
        valueElement.classList.toggle('is-missing', value == null);
        valueElement.parentElement.setAttribute('aria-label', `${model.name}: ${formattedValue}`);
      });
      labels.children[metricIndex].classList.add('is-hovered');
      tooltip.setAttribute('aria-hidden', 'false');
      positionTooltip(metricIndex);
    };

    const hideMetricTooltip = () => {
      if (activeMetricIndex >= 0) labels.children[activeMetricIndex].classList.remove('is-hovered');
      activeMetricIndex = -1;
      tooltip.setAttribute('aria-hidden', 'true');
    };

    labels.querySelectorAll('[data-radar-metric-index]').forEach((target) => {
      const index = Number(target.dataset.radarMetricIndex);
      target.addEventListener('pointerenter', () => showMetricTooltip(index));
      target.addEventListener('pointerleave', hideMetricTooltip);
    });

    scrollContainer.addEventListener('scroll', () => positionTooltip(activeMetricIndex), { passive: true });
    window.addEventListener('resize', () => positionTooltip(activeMetricIndex));

    const legendItems = [...radar.querySelectorAll('[data-radar-model]')];
    const legendPill = radar.querySelector('[data-radar-legend-pill]');
    let hoveredModel = '';
    let chartHoveredModel = '';
    let focusedModel = '';
    let selectedModel = '';
    let suppressedHoveredModel = '';
    let suppressedFocusedModel = '';
    let hoverClearTimer = 0;

    const positionLegendPill = (item) => {
      if (!item) {
        legendPill.classList.remove('is-visible');
        return;
      }

      if (legendPill.dataset.ready !== 'true') {
        legendPill.style.transition = 'none';
      }
      legendPill.style.width = `${item.offsetWidth}px`;
      legendPill.style.height = `${item.offsetHeight}px`;
      legendPill.style.transform = `translate3d(${item.offsetLeft}px, ${item.offsetTop}px, 0)`;
      legendPill.classList.add('is-visible');

      if (legendPill.dataset.ready !== 'true') {
        legendPill.getBoundingClientRect();
        legendPill.style.transition = '';
        legendPill.dataset.ready = 'true';
      }
    };

    const getActiveModel = () => {
      const hoverPreview = hoveredModel === suppressedHoveredModel ? '' : hoveredModel;
      const focusPreview = focusedModel === suppressedFocusedModel ? '' : focusedModel;
      return hoverPreview || chartHoveredModel || focusPreview || selectedModel;
    };

    const updateActiveModel = () => {
      const activeModel = getActiveModel();
      radar.classList.toggle('is-radar-filtered', Boolean(activeModel));
      if (activeModel) {
        radar.dataset.activeModel = activeModel;
      } else {
        delete radar.dataset.activeModel;
      }
      legendItems.forEach((item) => {
        item.classList.toggle('is-active', item.dataset.radarModel === activeModel);
        item.setAttribute('aria-pressed', String(item.dataset.radarModel === selectedModel));
      });
      positionLegendPill(legendItems.find((item) => item.dataset.radarModel === activeModel));
    };

    legendItems.forEach((item) => {
      item.setAttribute('aria-pressed', 'false');
      item.addEventListener('pointerenter', () => {
        window.clearTimeout(hoverClearTimer);
        if (suppressedHoveredModel && suppressedHoveredModel !== item.dataset.radarModel) {
          suppressedHoveredModel = '';
        }
        hoveredModel = item.dataset.radarModel;
        updateActiveModel();
      });
      item.addEventListener('pointerleave', () => {
        window.clearTimeout(hoverClearTimer);
        hoverClearTimer = window.setTimeout(() => {
          hoveredModel = '';
          if (suppressedHoveredModel === item.dataset.radarModel) suppressedHoveredModel = '';
          updateActiveModel();
        }, 90);
      });
      item.addEventListener('focus', () => {
        if (suppressedFocusedModel && suppressedFocusedModel !== item.dataset.radarModel) {
          suppressedFocusedModel = '';
        }
        focusedModel = item.dataset.radarModel;
        updateActiveModel();
      });
      item.addEventListener('blur', () => {
        focusedModel = '';
        if (suppressedFocusedModel === item.dataset.radarModel) suppressedFocusedModel = '';
        updateActiveModel();
      });
      item.addEventListener('click', () => {
        const model = item.dataset.radarModel;
        if (selectedModel === model) {
          selectedModel = '';
          suppressedHoveredModel = model;
          suppressedFocusedModel = model;
        } else {
          selectedModel = model;
          suppressedHoveredModel = '';
          suppressedFocusedModel = '';
        }
        updateActiveModel();
      });
    });

    const distanceToSegment = (cursor, from, to) => {
      const dx = to.x - from.x;
      const dy = to.y - from.y;
      const lengthSquared = dx * dx + dy * dy;
      const t = lengthSquared === 0
        ? 0
        : Math.max(0, Math.min(1, ((cursor.x - from.x) * dx + (cursor.y - from.y) * dy) / lengthSquared));
      const x = from.x + dx * t;
      const y = from.y + dy * t;
      return {
        distance: Math.hypot(cursor.x - x, cursor.y - y),
        score: from.value + (to.value - from.value) * t
      };
    };

    const hitTestModel = (event) => {
      if (event.pointerType === 'touch' || event.target.closest?.('.benchmark-radar-label-target')) return '';

      const matrix = svg.getScreenCTM();
      if (!matrix) return '';
      const cursor = new DOMPoint(event.clientX, event.clientY).matrixTransform(matrix.inverse());
      if (Math.hypot(cursor.x - center.x, cursor.y - center.y) > radius + 14) return '';

      const scaleX = Math.hypot(matrix.a, matrix.b);
      const scaleY = Math.hypot(matrix.c, matrix.d);
      const screenScale = Math.max(Math.min(scaleX, scaleY), 0.001);
      const hitDistance = 10 / screenScale;
      const ambiguityDistance = 4.5 / screenScale;
      const hits = modelGeometry.map(({ model, points, segments }) => {
        let distance = Infinity;
        let score = -Infinity;
        const consider = (nextDistance, nextScore) => {
          if (nextDistance < distance - 0.01 || (Math.abs(nextDistance - distance) <= 0.01 && nextScore > score)) {
            distance = nextDistance;
            score = nextScore;
          }
        };

        points.forEach((mark) => {
          consider(Math.hypot(cursor.x - mark.x, cursor.y - mark.y), mark.value);
        });
        segments.forEach(({ from, to }) => {
          const nearest = distanceToSegment(cursor, from, to);
          consider(nearest.distance, nearest.score);
        });
        return { slug: model.slug, distance, score };
      }).filter((hit) => hit.distance <= hitDistance);

      if (hits.length === 0) return '';
      const nearestDistance = Math.min(...hits.map((hit) => hit.distance));
      const ambiguousHits = hits.filter((hit) => hit.distance <= nearestDistance + ambiguityDistance);
      const inkling = ambiguousHits.find((hit) => hit.slug === 'inkling');
      if (inkling) return inkling.slug;

      ambiguousHits.sort((a, b) => a.score - b.score || a.distance - b.distance);
      return ambiguousHits[0].slug;
    };

    const setChartHoveredModel = (model) => {
      if (chartHoveredModel === model) return;
      chartHoveredModel = model;
      updateActiveModel();
    };

    svg.addEventListener('pointermove', (event) => {
      setChartHoveredModel(hitTestModel(event));
    });
    svg.addEventListener('pointerleave', () => setChartHoveredModel(''));
    svg.addEventListener('pointercancel', () => setChartHoveredModel(''));

    window.addEventListener('resize', () => {
      const activeModel = getActiveModel();
      positionLegendPill(legendItems.find((item) => item.dataset.radarModel === activeModel));
    });

    radar.classList.add('is-radar-pending');
    const reveal = () => {
      radar.classList.remove('is-radar-pending');
      radar.classList.add('is-radar-visible');
    };

    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      reveal();
    } else if ('IntersectionObserver' in window) {
      const observer = new IntersectionObserver((entries) => {
        if (!entries.some((entry) => entry.isIntersecting)) return;
        observer.disconnect();
        reveal();
      }, { threshold: 0.18, rootMargin: '0px 0px -10% 0px' });
      observer.observe(radar);
    } else {
      reveal();
    }
    };

    /* Geometry construction creates more than one hundred SVG nodes and calls
       getBBox(). Defer that work until the below-the-fold chart is close to the
       viewport on both Inkling pages. */
    if ('IntersectionObserver' in window) {
      const initializeObserver = new IntersectionObserver((entries) => {
        if (!entries.some((entry) => entry.isIntersecting)) return;
        initializeObserver.disconnect();
        initialize();
      }, { threshold: 0.01, rootMargin: '200px 0px' });
      initializeObserver.observe(radar);
    } else {
      initialize();
    }
  })();
})();</script></div><div class="demo-panel spider-switch-panel" id="spider-switch-79d4c5f4-panel-small" role="tabpanel" aria-labelledby="spider-switch-79d4c5f4-tab-small" hidden><style>  .benchmark-radar {
    --radar-inkling: #0155BF;

    --radar-inklingsmall: #8f3f71;

    --radar-nemotron: #EF402C;
    --radar-glm: #21B3FF;
    --radar-gpt: #F7A224;
    --radar-claude: #0E9972;

    /* Family-chart comparison set reuses the comparison hues above, so the
       two variants of this chart stay one palette. */
    --radar-deepseek: #21B3FF;
    --radar-gemini: #F7A224;
    --radar-luna: #0E9972;

    --radar-grid: color-mix(in srgb, var(--fg4) 17%, var(--bg) 83%);
    --radar-axis: color-mix(in srgb, var(--fg4) 25%, var(--bg) 75%);
    color: var(--fg);
    font-family: var(--font-sans);
    margin: 0 auto;
    max-width: 860px;
    position: relative;
    width: 100%;
  }

  @media (min-width: 1171px) {
    #main article.content .benchmark-radar-figure {
        margin-left: calc((100% - 860px) / 2);
      margin-right: 0;
      max-width: none;
        width: 860px;
    }
  }

  #main article.content .benchmark-radar-legend {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    font-size: 15px;
    gap: 0.35rem 0.3rem;
    justify-content: center;
    margin: 0.7rem auto 0;
    position: relative;
  }

  #main article.content .benchmark-radar-legend-pill {
    background: color-mix(in srgb, var(--fg) 10%, var(--bg) 90%);
    border-radius: 999px;
    height: 0;
    left: 0;
    opacity: 0;
    pointer-events: none;
    position: absolute;
    top: 0;
    transform: translate3d(0, 0, 0);
    transition:
      opacity 120ms ease,
      transform 280ms cubic-bezier(0.22, 1, 0.36, 1),
      width 240ms cubic-bezier(0.22, 1, 0.36, 1),
      height 240ms cubic-bezier(0.22, 1, 0.36, 1);
    width: 0;
    will-change: transform, width, height;
    z-index: 0;
  }

  #main article.content .benchmark-radar-legend-pill.is-visible {
    opacity: 1;
  }

  #main article.content .benchmark-radar-legend-item {
    -webkit-appearance: none;
    align-items: center;
    appearance: none;
    background: transparent;
    border: 0;
    border-radius: 999px;
    color: var(--fg);
    cursor: pointer;
    display: inline-flex;
    font: inherit;
    font-weight: 400;
    gap: 0.28rem;
    line-height: 1;
    margin: 0;
    padding: 0.62rem 0.9rem;
    position: relative;
    white-space: nowrap;
    z-index: 1;
  }

  html.js #main article.content .benchmark-radar-legend-item:hover,
  html.js #main article.content .benchmark-radar-legend-item:focus-visible {
    background: transparent;
  }

  #main article.content .benchmark-radar-legend-item:focus-visible {
    outline: 2px solid color-mix(in srgb, var(--fg) 42%, transparent 58%);
    outline-offset: 2px;
  }

  html:not(.js) #main article.content .benchmark-radar-legend-item:hover,
  html:not(.js) #main article.content .benchmark-radar-legend-item:focus-visible {
    background: color-mix(in srgb, var(--fg4) 11%, var(--bg) 89%);
  }

  .benchmark-radar-key {
    border-radius: 50%;
    box-sizing: border-box;
    display: inline-block;
    flex: 0 0 auto;
    height: 9px;
    width: 9px;
  }

  /* Every key is a hollow ring in its model's colour; the hero's is filled in.
     Keyed on --hero rather than on a slug so the landing page's two charts,
     whose styles share one document, cannot fill each other's keys. */
  .benchmark-radar-key {
    background: transparent;
    border: 2px solid transparent;
  }
  .benchmark-radar-key.benchmark-radar-key--inkling { border-color: var(--radar-inkling); }
  .benchmark-radar-key.benchmark-radar-key--inklingsmall { border-color: var(--radar-inklingsmall); }
  .benchmark-radar-key--hero.benchmark-radar-key--inkling {
    background: var(--radar-inkling);
    border-color: transparent;
  }
  .benchmark-radar-key--hero.benchmark-radar-key--inklingsmall {
    background: var(--radar-inklingsmall);
    border-color: transparent;
  }
  .benchmark-radar-key.benchmark-radar-key--nemotron { border-color: var(--radar-nemotron); }
  .benchmark-radar-key.benchmark-radar-key--glm { border-color: var(--radar-glm); }
  .benchmark-radar-key.benchmark-radar-key--gpt { border-color: var(--radar-gpt); }
  .benchmark-radar-key.benchmark-radar-key--claude { border-color: var(--radar-claude); }

  .benchmark-radar-key.benchmark-radar-key--deepseek { border-color: var(--radar-deepseek); }
  .benchmark-radar-key.benchmark-radar-key--gemini { border-color: var(--radar-gemini); }
  .benchmark-radar-key.benchmark-radar-key--luna { border-color: var(--radar-luna); }



  /* Inkling-Small is ours, so it reads like Inkling — solid line, filled key and
     points — against the dashed, hollow comparison set. Inkling keeps the heaviest
     stroke and the largest markers, so it stays the hero of the pair. These rules
     follow the `:not(--inkling)` defaults above and win on source order at equal
     specificity. */
  .benchmark-radar-series--inklingsmall .benchmark-radar-segment { stroke-width: 2.4; }

  .benchmark-radar-series--hero.benchmark-radar-series--inklingsmall .benchmark-radar-point {
    fill: var(--radar-inklingsmall);
    stroke: var(--bg) !important;
    stroke-width: 1.5;
  }

  .benchmark-radar.is-radar-filtered[data-active-model="inklingsmall"] .benchmark-radar-series--inklingsmall {
    opacity: 1 !important;
  }

  .benchmark-radar.is-radar-filtered[data-active-model="inklingsmall"] .benchmark-radar-series--inklingsmall .benchmark-radar-area {
    fill-opacity: 0.2;
    opacity: 1 !important;
  }


  .benchmark-radar-scroll {
    display: block;
    max-width: 100%;
    min-width: 0;
    overflow-x: auto;
    overflow-y: hidden;
    overscroll-behavior-x: contain;
    width: 100%;
    -webkit-overflow-scrolling: touch;
  }

  .benchmark-radar-scroll::-webkit-scrollbar { height: 6px; }
  .benchmark-radar-scroll::-webkit-scrollbar-thumb {
    background: color-mix(in srgb, var(--fg4) 45%, var(--bg) 55%);
    border-radius: 3px;
  }

  .benchmark-radar-svg {
    min-width: 720px;
    overflow: visible;
  }

  .benchmark-radar-svg text {
    fill: var(--fg2);
    font-family: var(--font-sans);
    font-size: 13px;
  }

  .benchmark-radar-svg .benchmark-radar-tick {
    fill: var(--fg4);
    font-size: 13px;
  }

  .benchmark-radar-svg .benchmark-radar-tick--detail {
    opacity: 0;
    transition: opacity 160ms ease;
  }

  .benchmark-radar.is-radar-filtered .benchmark-radar-tick--detail {
    opacity: 1;
  }

  .benchmark-radar-label {
    pointer-events: none;
    transform-box: fill-box;
    transform-origin: center;
    transition: fill 140ms ease, transform 180ms cubic-bezier(0.22, 1, 0.36, 1);
  }

  .benchmark-radar-label-title {
    fill: var(--fg);
  }

  .benchmark-radar-label-category {
    fill: var(--fg4);
    opacity: 0.82;
  }

  .benchmark-radar-label-target {
    cursor: default;
  }

  .benchmark-radar-label-hit {
    fill: transparent;
    pointer-events: all;
  }

  .benchmark-radar-label-target.is-hovered .benchmark-radar-label {
    fill: var(--fg);
    transform: scale(1.16);
  }

  .benchmark-radar-grid {
    fill: none;
    stroke: var(--radar-grid);
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
  }

  .benchmark-radar-axis {
    stroke: var(--radar-axis);
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
  }

  .benchmark-radar-area { pointer-events: none; }
  .benchmark-radar-segment,
  .benchmark-radar-point {
    vector-effect: non-scaling-stroke;
  }

  .benchmark-radar-segment {
    fill: none;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-width: 2;
  }

  .benchmark-radar-series--inkling .benchmark-radar-segment { stroke-width: 2.8; }
  .benchmark-radar-point {
    fill: var(--bg);
    stroke-width: 2.2;
  }

  .benchmark-radar-series--hero.benchmark-radar-series--inkling .benchmark-radar-point {
    fill: var(--radar-inkling);
    stroke: var(--bg) !important;
    stroke-width: 1.5;
  }

  .benchmark-radar-series {
    transition: opacity 180ms ease;
    transform-box: view-box;
    transform-origin: 50% 50%;
  }

  .benchmark-radar-area {
    transition: fill-opacity 180ms ease, opacity 180ms ease;
  }

  .benchmark-radar.is-radar-filtered .benchmark-radar-series {
    opacity: 0.18 !important;
  }

  .benchmark-radar.is-radar-filtered[data-active-model="inkling"] .benchmark-radar-series--inkling,
  .benchmark-radar.is-radar-filtered[data-active-model="nemotron"] .benchmark-radar-series--nemotron,
  .benchmark-radar.is-radar-filtered[data-active-model="glm"] .benchmark-radar-series--glm,
  .benchmark-radar.is-radar-filtered[data-active-model="gpt"] .benchmark-radar-series--gpt,
  .benchmark-radar.is-radar-filtered[data-active-model="claude"] .benchmark-radar-series--claude {
    opacity: 1 !important;
  }


  .benchmark-radar.is-radar-filtered[data-active-model="deepseek"] .benchmark-radar-series--deepseek,
  .benchmark-radar.is-radar-filtered[data-active-model="gemini"] .benchmark-radar-series--gemini,
  .benchmark-radar.is-radar-filtered[data-active-model="luna"] .benchmark-radar-series--luna {
    opacity: 1 !important;
  }


  .benchmark-radar.is-radar-filtered .benchmark-radar-area {
    opacity: 0 !important;
  }

  .benchmark-radar.is-radar-filtered[data-active-model="inkling"] .benchmark-radar-series--inkling .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="nemotron"] .benchmark-radar-series--nemotron .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="glm"] .benchmark-radar-series--glm .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="gpt"] .benchmark-radar-series--gpt .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="claude"] .benchmark-radar-series--claude .benchmark-radar-area {
    fill-opacity: 0.2;
    opacity: 1 !important;
  }


  .benchmark-radar.is-radar-filtered[data-active-model="deepseek"] .benchmark-radar-series--deepseek .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="gemini"] .benchmark-radar-series--gemini .benchmark-radar-area,
  .benchmark-radar.is-radar-filtered[data-active-model="luna"] .benchmark-radar-series--luna .benchmark-radar-area {
    fill-opacity: 0.2;
    opacity: 1 !important;
  }


  .benchmark-radar.is-radar-filtered[data-active-model="inkling"] .benchmark-radar-series--inkling .benchmark-radar-area {
    fill-opacity: 0.26;
  }

  #main article.content .benchmark-radar-tooltip {
    background: #111;
    border: 0;
    border-radius: 4px;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.16);
    color: #fff;
    font-family: var(--font-sans);
    font-size: 11px;
    font-weight: 400;
    left: 0;
    line-height: 1;
    max-width: calc(100% - 16px);
    min-width: 190px;
    opacity: 0;
    padding: 0.45rem 0.55rem;
    pointer-events: none;
    position: absolute;
    top: 0;
    transform: translateY(4px);
    transition: opacity 140ms ease, transform 140ms ease;
    width: 210px;
    z-index: 3;
  }

  #main article.content .benchmark-radar-tooltip[aria-hidden="false"] {
    opacity: 1;
    transform: translateY(0);
  }

  #main article.content .benchmark-radar-tooltip-title {
    color: #fff;
    font-weight: 500;
    margin-bottom: 0.4rem;
  }

  #main article.content .benchmark-radar-tooltip [data-radar-tooltip-rows] {
    display: grid;
    gap: 0.15rem;
    margin: 0;
  }

  #main article.content .benchmark-radar-tooltip-row {
    align-items: center;
    display: grid;
    gap: 0.4rem;
    grid-template-columns: 8px minmax(0, 1fr) auto;
    line-height: 1.35;
    margin: 0;
  }

  #main article.content .benchmark-radar-tooltip-dot {
    border-radius: 50%;
    height: 8px;
    width: 8px;
  }

  #main article.content .benchmark-radar-tooltip-value {
    font-variant-numeric: tabular-nums;
    text-align: right;
  }

  #main article.content .benchmark-radar-tooltip-value.is-missing {
    color: rgba(255, 255, 255, 0.62);
  }

  .benchmark-radar-series--hero .benchmark-radar-point {
    transform-box: fill-box;
    transform-origin: center;
  }

  html.js .benchmark-radar.is-radar-pending .benchmark-radar-series--hero .benchmark-radar-point {
    opacity: 0;
    transform: scale(0);
  }

  html.js .benchmark-radar.is-radar-pending .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment) {
    opacity: 0;
  }

  html.js .benchmark-radar.is-radar-pending .benchmark-radar-series:not(.benchmark-radar-series--hero) {
    opacity: 0;
    transform: scale(0.99);
  }

  html.js .benchmark-radar.is-radar-visible .benchmark-radar-series--hero .benchmark-radar-point {
    animation: benchmark-radar-point-in 260ms cubic-bezier(0.2, 0.78, 0.22, 1) both;
    animation-delay: var(--radar-point-delay, 0ms);
  }

  html.js .benchmark-radar.is-radar-visible .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment) {
    animation: benchmark-radar-hero-shape-in 320ms ease-out 820ms both;
  }

  html.js .benchmark-radar.is-radar-visible .benchmark-radar-series:not(.benchmark-radar-series--hero) {
    animation: benchmark-radar-comparison-in 480ms ease-out 820ms both;
  }

  @keyframes benchmark-radar-point-in {
    0% { opacity: 0; transform: scale(0); }
    70% { opacity: 1; transform: scale(1.2); }
    100% { opacity: 1; transform: scale(1); }
  }

  @keyframes benchmark-radar-hero-shape-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes benchmark-radar-comparison-in {
    from { opacity: 0; transform: scale(0.99); }
    to { opacity: 1; transform: scale(1); }
  }

  @media (prefers-reduced-motion: reduce) {
    #main article.content .benchmark-radar-legend-pill {
      transition: none;
    }

    .benchmark-radar-label {
      transition: none;
    }

    html.js .benchmark-radar.is-radar-pending .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment, .benchmark-radar-point),
    html.js .benchmark-radar.is-radar-visible .benchmark-radar-series--hero :is(.benchmark-radar-area, .benchmark-radar-segment, .benchmark-radar-point),
    html.js .benchmark-radar.is-radar-pending .benchmark-radar-series:not(.benchmark-radar-series--hero),
    html.js .benchmark-radar.is-radar-visible .benchmark-radar-series:not(.benchmark-radar-series--hero) {
      animation: none;
      opacity: 1;
      transform: none;
    }
  }

  @media (max-width: 620px) {
    #main article.content .benchmark-radar-legend {
      justify-content: center;
    }

    /* Phones: scale the whole chart to fit instead of horizontal-scrolling a
       radar (you can't read a shape you can only see half of). Labels are
       enlarged in user units to stay legible once the SVG is scaled down. */
    .benchmark-radar-scroll {
      overflow-x: visible;
    }

    .benchmark-radar-svg {
      min-width: 0;
      width: 100%;
      height: auto;
    }

    .benchmark-radar-svg text {
      font-size: 21px;
    }

    .benchmark-radar-svg .benchmark-radar-tick {
      font-size: 19px;
    }
  }</style><div class="benchmark-radar-figure"><div class="benchmark-radar" id="benchmark-radar-small"><div class="benchmark-radar-scroll"><svg class="benchmark-radar-svg" data-benchmark-radar viewBox="0 0 860 600" role="img" aria-label="Inkling benchmark comparison" aria-describedby="benchmark-radar-desc-small"><desc id="benchmark-radar-desc-small">Spider chart comparing Inkling-Small, Inkling, DeepSeek V4 Flash, Gemini 3.5 Flash-Lite, and GPT 5.6 Luna on ten evaluations scored from zero to one hundred. Inkling-Small is shown with a solid purple line, Inkling with the heavier cobalt line, and the comparison models with dashed lines. Evaluations without a reported model score are plotted at zero. Hover an evaluation to compare every model&#39;s score.</desc><g data-radar-grid></g><g data-radar-axes></g><g data-radar-series></g><g data-radar-labels></g></svg></div><div class="benchmark-radar-legend" aria-label="Model legend"><span class="benchmark-radar-legend-pill" data-radar-legend-pill aria-hidden="true"></span><button class="benchmark-radar-legend-item" type="button" data-radar-model="inklingsmall"><span class="benchmark-radar-key benchmark-radar-key--inklingsmall benchmark-radar-key--hero" aria-hidden="true"></span>Inkling-Small</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="inkling"><span class="benchmark-radar-key benchmark-radar-key--inkling" aria-hidden="true"></span>Inkling</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="deepseek"><span class="benchmark-radar-key benchmark-radar-key--deepseek" aria-hidden="true"></span>DeepSeek V4 Flash</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="gemini"><span class="benchmark-radar-key benchmark-radar-key--gemini" aria-hidden="true"></span>Gemini 3.5 Flash-Lite</button><button class="benchmark-radar-legend-item" type="button" data-radar-model="luna"><span class="benchmark-radar-key benchmark-radar-key--luna" aria-hidden="true"></span>GPT 5.6 Luna</button></div><div class="benchmark-radar-tooltip tooltip" data-radar-tooltip role="tooltip" aria-hidden="true"><div class="benchmark-radar-tooltip-title" data-radar-tooltip-title></div><div data-radar-tooltip-rows></div></div></div></div><script>(function(){const variant = "small";
const hero = "inklingsmall";
const metrics = [{"name":"HLE (no tools)","category":"Reasoning"},{"name":"AIME 2026","category":"Reasoning"},{"name":"SWE-bench Pro (Public)","category":"Agentic (coding)"},{"name":"Terminal Bench 2.1","category":"Agentic (coding)"},{"name":"MCP Atlas (public)","category":"Agentic (general)"},{"name":"SimpleQA Verified","category":"Factuality"},{"name":"AA Omniscience (acc)","category":"Factuality"},{"name":"AudioMC","category":"Audio"},{"name":"MMMU Pro","category":"Vision"},{"name":"IFBench","category":"Chat"}];
const models = [{"name":"Inkling-Small","slug":"inklingsmall","color":"var(--radar-inklingsmall)","dash":"","fillOpacity":0.12,"values":[31.6,95.5,55.9,64.7,79.6,20.6,30.5,54.9,74,82.2]},{"name":"Inkling","slug":"inkling","color":"var(--radar-inkling)","dash":"12 4","fillOpacity":0.035,"values":[29.7,97.1,54.3,63.8,78.8,43.9,40,56.6,73.5,79.8]},{"name":"DeepSeek V4 Flash","slug":"deepseek","color":"var(--radar-deepseek)","dash":"2 4","fillOpacity":0.035,"values":[32.1,95.8,52.6,61.8,69,34.1,37.2,null,null,79.2]},{"name":"Gemini 3.5 Flash-Lite","slug":"gemini","color":"var(--radar-gemini)","dash":"9 4 2 4","fillOpacity":0.035,"values":[17.5,82.2,54.2,54,79.8,44.1,30.3,33.6,79,78.6]},{"name":"GPT 5.6 Luna","slug":"luna","color":"var(--radar-luna)","dash":"6 4","fillOpacity":0.035,"values":[35.6,97.6,62.7,82.5,77,41.7,41.3,null,78.6,67.3]}];

  (() => {
    const radar = document.getElementById('benchmark-radar-' + variant);
    const heroSlug = hero;
    const svg = radar && radar.querySelector('[data-benchmark-radar]');
    if (!radar || !svg) return;

    const initialize = () => {
      if (svg.dataset.ready === 'true') return;
      svg.dataset.ready = 'true';

    const ns = 'http://www.w3.org/2000/svg';
    const center = { x: 430, y: 300 };
    const radius = 220;

    const create = (tag, attrs = {}) => {
      const node = document.createElementNS(ns, tag);
      Object.entries(attrs).forEach(([key, value]) => node.setAttribute(key, value));
      return node;
    };
    const point = (index, value, extra = 0) => {
      const angle = -Math.PI / 2 + index * Math.PI * 2 / metrics.length;
      const distance = radius * value / 100 + extra;
      return {
        x: center.x + Math.cos(angle) * distance,
        y: center.y + Math.sin(angle) * distance,
        ux: Math.cos(angle),
        uy: Math.sin(angle)
      };
    };
    const modelGeometry = models.map((model) => {
      const points = model.values.map((value, index) => {
        const score = value ?? 0;
        return { index, value: score, ...point(index, score) };
      });
      const segments = points.map((from, pointIndex) => ({
        from,
        to: points[(pointIndex + 1) % points.length]
      }));
      return { model, points, segments };
    });
    const pointsAttr = (value) => metrics.map((_, index) => {
      const p = point(index, value);
      return `${p.x.toFixed(2)},${p.y.toFixed(2)}`;
    }).join(' ');

    const grid = svg.querySelector('[data-radar-grid]');
    [20, 40, 60, 80, 100].forEach((value) => {
      grid.append(create('polygon', { class: 'benchmark-radar-grid', points: pointsAttr(value) }));
      const p = point(0, value);
      const tick = create('text', {
        class: 'benchmark-radar-tick benchmark-radar-tick--detail',
        x: p.x + 7,
        y: p.y + 16
      });
      tick.textContent = value;
      grid.append(tick);
    });

    const axes = svg.querySelector('[data-radar-axes]');
    metrics.forEach((_, index) => {
      const p = point(index, 100);
      axes.append(create('line', {
        class: 'benchmark-radar-axis',
        x1: center.x, y1: center.y, x2: p.x, y2: p.y
      }));
    });

    const seriesLayer = svg.querySelector('[data-radar-series]');
    [...modelGeometry].reverse().forEach(({ model, points }) => {
      const group = create('g', {
        class: `benchmark-radar-series benchmark-radar-series--${model.slug}` + (model.slug === heroSlug ? ' benchmark-radar-series--hero' : '')
      });

      const areaPath = points.map((p, index) => `${index === 0 ? 'M' : 'L'} ${p.x.toFixed(2)} ${p.y.toFixed(2)}`).join(' ') + ' Z';
      group.append(create('path', {
        class: 'benchmark-radar-area',
        d: areaPath,
        fill: model.color,
        'fill-opacity': model.fillOpacity
      }));

      points.forEach((from, pointIndex) => {
        const to = points[(pointIndex + 1) % points.length];
        const segment = create('line', {
          class: 'benchmark-radar-segment',
          x1: from.x, y1: from.y, x2: to.x, y2: to.y,
          stroke: model.color
        });
        if (model.dash) segment.setAttribute('stroke-dasharray', model.dash);
        group.append(segment);
      });

      points.forEach((p) => {
        const marker = create('circle', {
          class: 'benchmark-radar-point',
          cx: p.x, cy: p.y, r: model.slug === heroSlug ? 5 : 4.5,
          stroke: model.color,
          style: `--radar-point-delay: ${p.index * 55}ms`
        });
        group.append(marker);
      });
      seriesLayer.append(group);
    });

    const labels = svg.querySelector('[data-radar-labels]');
    metrics.forEach((metric, index) => {
      const p = point(index, 100, 28);
      const target = create('g', {
        class: 'benchmark-radar-label-target',
        'data-radar-metric-index': index
      });
      const label = create('text', { class: 'benchmark-radar-label', x: p.x, y: p.y + 5 });
      label.setAttribute('text-anchor', p.ux > 0.24 ? 'start' : p.ux < -0.24 ? 'end' : 'middle');
      const title = create('tspan', { class: 'benchmark-radar-label-title', x: p.x, dy: 0 });
      title.textContent = metric.name;
      const category = create('tspan', { class: 'benchmark-radar-label-category', x: p.x, dy: '1.3em' });
      category.textContent = metric.category;
      label.append(title, category);
      target.append(label);
      labels.append(target);

      /* A long name on a side axis can run past the SVG box once the phone
         breakpoint enlarges label text in user units — "SWE-bench Pro (Public)"
         overhung the right edge by ~30 units at 390px wide. Only when a label
         actually overflows, break its trailing parenthetical onto a second line.
         Measuring rather than hard-coding the one name keeps the qualifier and
         the enlarged font, leaves every label that already fits alone, and
         re-evaluates itself if names, font sizes or breakpoints change later.
         Note Terminal Bench 2.1 shares this side with only ~20 units to spare,
         so shifting the labels outward is not an option. */
      const box = svg.viewBox.baseVal;
      const first = label.getBBox();
      const parenthetical = metric.name.match(/^(.*\S)\s+(\(.*\))$/);
      if (parenthetical && (first.x < box.x || first.x + first.width > box.x + box.width)) {
        title.textContent = parenthetical[1];
        const rest = create('tspan', { class: 'benchmark-radar-label-title', x: p.x, dy: '1.15em' });
        rest.textContent = parenthetical[2];
        label.insertBefore(rest, category);
      }

      const bounds = label.getBBox();
      target.insertBefore(create('rect', {
        class: 'benchmark-radar-label-hit',
        x: bounds.x - 8,
        y: bounds.y - 6,
        width: bounds.width + 16,
        height: bounds.height + 12,
        rx: 4
      }), label);
    });

    const tooltip = radar.querySelector('[data-radar-tooltip]');
    const tooltipTitle = tooltip.querySelector('[data-radar-tooltip-title]');
    const tooltipRows = tooltip.querySelector('[data-radar-tooltip-rows]');
    const scrollContainer = radar.querySelector('.benchmark-radar-scroll');
    let activeMetricIndex = -1;

    const formatTooltipValue = (metric, value) => {
      if (value != null) return value.toFixed(1);
      if (metric.category === 'Audio') return 'audio not supported';
      if (metric.category === 'Vision') return 'vision not supported';
      return '0.0 - not reported';
    };

    models.forEach((model) => {
      const row = document.createElement('div');
      row.className = 'benchmark-radar-tooltip-row';
      row.dataset.radarTooltipModel = model.slug;

      const dot = document.createElement('span');
      dot.className = 'benchmark-radar-tooltip-dot';
      dot.style.background = model.color;
      dot.setAttribute('aria-hidden', 'true');

      const name = document.createElement('span');
      name.textContent = model.name;

      const value = document.createElement('span');
      value.className = 'benchmark-radar-tooltip-value';
      value.dataset.radarTooltipValue = '';

      row.append(dot, name, value);
      tooltipRows.append(row);
    });

    const positionTooltip = (metricIndex) => {
      if (metricIndex < 0) return;
      const rootRect = radar.getBoundingClientRect();
      const svgRect = svg.getBoundingClientRect();
      const viewBox = svg.viewBox.baseVal;
      const anchor = point(metricIndex, 100, 24);
      const anchorX = svgRect.left - rootRect.left + (anchor.x - viewBox.x) * svgRect.width / viewBox.width;
      const anchorY = svgRect.top - rootRect.top + (anchor.y - viewBox.y) * svgRect.height / viewBox.height;
      const gap = 12;
      const inset = 8;
      const tooltipWidth = tooltip.offsetWidth;
      const tooltipHeight = tooltip.offsetHeight;
      const rootWidth = radar.clientWidth;
      const rootHeight = radar.clientHeight;
      const placeBelow = anchorY < rootHeight * 0.48;
      const left = Math.min(Math.max(anchorX - tooltipWidth / 2, inset), rootWidth - tooltipWidth - inset);
      const preferredTop = placeBelow ? anchorY + gap : anchorY - tooltipHeight - gap;
      const top = Math.min(Math.max(preferredTop, inset), rootHeight - tooltipHeight - inset);
      tooltip.style.left = `${left}px`;
      tooltip.style.top = `${top}px`;
    };

    const showMetricTooltip = (metricIndex) => {
      activeMetricIndex = metricIndex;
      tooltipTitle.textContent = metrics[metricIndex].name;
      models.forEach((model) => {
        const value = model.values[metricIndex];
        const formattedValue = formatTooltipValue(metrics[metricIndex], value);
        const valueElement = tooltip.querySelector(`[data-radar-tooltip-model="${model.slug}"] [data-radar-tooltip-value]`);
        valueElement.textContent = formattedValue;
        valueElement.classList.toggle('is-missing', value == null);
        valueElement.parentElement.setAttribute('aria-label', `${model.name}: ${formattedValue}`);
      });
      labels.children[metricIndex].classList.add('is-hovered');
      tooltip.setAttribute('aria-hidden', 'false');
      positionTooltip(metricIndex);
    };

    const hideMetricTooltip = () => {
      if (activeMetricIndex >= 0) labels.children[activeMetricIndex].classList.remove('is-hovered');
      activeMetricIndex = -1;
      tooltip.setAttribute('aria-hidden', 'true');
    };

    labels.querySelectorAll('[data-radar-metric-index]').forEach((target) => {
      const index = Number(target.dataset.radarMetricIndex);
      target.addEventListener('pointerenter', () => showMetricTooltip(index));
      target.addEventListener('pointerleave', hideMetricTooltip);
    });

    scrollContainer.addEventListener('scroll', () => positionTooltip(activeMetricIndex), { passive: true });
    window.addEventListener('resize', () => positionTooltip(activeMetricIndex));

    const legendItems = [...radar.querySelectorAll('[data-radar-model]')];
    const legendPill = radar.querySelector('[data-radar-legend-pill]');
    let hoveredModel = '';
    let chartHoveredModel = '';
    let focusedModel = '';
    let selectedModel = '';
    let suppressedHoveredModel = '';
    let suppressedFocusedModel = '';
    let hoverClearTimer = 0;

    const positionLegendPill = (item) => {
      if (!item) {
        legendPill.classList.remove('is-visible');
        return;
      }

      if (legendPill.dataset.ready !== 'true') {
        legendPill.style.transition = 'none';
      }
      legendPill.style.width = `${item.offsetWidth}px`;
      legendPill.style.height = `${item.offsetHeight}px`;
      legendPill.style.transform = `translate3d(${item.offsetLeft}px, ${item.offsetTop}px, 0)`;
      legendPill.classList.add('is-visible');

      if (legendPill.dataset.ready !== 'true') {
        legendPill.getBoundingClientRect();
        legendPill.style.transition = '';
        legendPill.dataset.ready = 'true';
      }
    };

    const getActiveModel = () => {
      const hoverPreview = hoveredModel === suppressedHoveredModel ? '' : hoveredModel;
      const focusPreview = focusedModel === suppressedFocusedModel ? '' : focusedModel;
      return hoverPreview || chartHoveredModel || focusPreview || selectedModel;
    };

    const updateActiveModel = () => {
      const activeModel = getActiveModel();
      radar.classList.toggle('is-radar-filtered', Boolean(activeModel));
      if (activeModel) {
        radar.dataset.activeModel = activeModel;
      } else {
        delete radar.dataset.activeModel;
      }
      legendItems.forEach((item) => {
        item.classList.toggle('is-active', item.dataset.radarModel === activeModel);
        item.setAttribute('aria-pressed', String(item.dataset.radarModel === selectedModel));
      });
      positionLegendPill(legendItems.find((item) => item.dataset.radarModel === activeModel));
    };

    legendItems.forEach((item) => {
      item.setAttribute('aria-pressed', 'false');
      item.addEventListener('pointerenter', () => {
        window.clearTimeout(hoverClearTimer);
        if (suppressedHoveredModel && suppressedHoveredModel !== item.dataset.radarModel) {
          suppressedHoveredModel = '';
        }
        hoveredModel = item.dataset.radarModel;
        updateActiveModel();
      });
      item.addEventListener('pointerleave', () => {
        window.clearTimeout(hoverClearTimer);
        hoverClearTimer = window.setTimeout(() => {
          hoveredModel = '';
          if (suppressedHoveredModel === item.dataset.radarModel) suppressedHoveredModel = '';
          updateActiveModel();
        }, 90);
      });
      item.addEventListener('focus', () => {
        if (suppressedFocusedModel && suppressedFocusedModel !== item.dataset.radarModel) {
          suppressedFocusedModel = '';
        }
        focusedModel = item.dataset.radarModel;
        updateActiveModel();
      });
      item.addEventListener('blur', () => {
        focusedModel = '';
        if (suppressedFocusedModel === item.dataset.radarModel) suppressedFocusedModel = '';
        updateActiveModel();
      });
      item.addEventListener('click', () => {
        const model = item.dataset.radarModel;
        if (selectedModel === model) {
          selectedModel = '';
          suppressedHoveredModel = model;
          suppressedFocusedModel = model;
        } else {
          selectedModel = model;
          suppressedHoveredModel = '';
          suppressedFocusedModel = '';
        }
        updateActiveModel();
      });
    });

    const distanceToSegment = (cursor, from, to) => {
      const dx = to.x - from.x;
      const dy = to.y - from.y;
      const lengthSquared = dx * dx + dy * dy;
      const t = lengthSquared === 0
        ? 0
        : Math.max(0, Math.min(1, ((cursor.x - from.x) * dx + (cursor.y - from.y) * dy) / lengthSquared));
      const x = from.x + dx * t;
      const y = from.y + dy * t;
      return {
        distance: Math.hypot(cursor.x - x, cursor.y - y),
        score: from.value + (to.value - from.value) * t
      };
    };

    const hitTestModel = (event) => {
      if (event.pointerType === 'touch' || event.target.closest?.('.benchmark-radar-label-target')) return '';

      const matrix = svg.getScreenCTM();
      if (!matrix) return '';
      const cursor = new DOMPoint(event.clientX, event.clientY).matrixTransform(matrix.inverse());
      if (Math.hypot(cursor.x - center.x, cursor.y - center.y) > radius + 14) return '';

      const scaleX = Math.hypot(matrix.a, matrix.b);
      const scaleY = Math.hypot(matrix.c, matrix.d);
      const screenScale = Math.max(Math.min(scaleX, scaleY), 0.001);
      const hitDistance = 10 / screenScale;
      const ambiguityDistance = 4.5 / screenScale;
      const hits = modelGeometry.map(({ model, points, segments }) => {
        let distance = Infinity;
        let score = -Infinity;
        const consider = (nextDistance, nextScore) => {
          if (nextDistance < distance - 0.01 || (Math.abs(nextDistance - distance) <= 0.01 && nextScore > score)) {
            distance = nextDistance;
            score = nextScore;
          }
        };

        points.forEach((mark) => {
          consider(Math.hypot(cursor.x - mark.x, cursor.y - mark.y), mark.value);
        });
        segments.forEach(({ from, to }) => {
          const nearest = distanceToSegment(cursor, from, to);
          consider(nearest.distance, nearest.score);
        });
        return { slug: model.slug, distance, score };
      }).filter((hit) => hit.distance <= hitDistance);

      if (hits.length === 0) return '';
      const nearestDistance = Math.min(...hits.map((hit) => hit.distance));
      const ambiguousHits = hits.filter((hit) => hit.distance <= nearestDistance + ambiguityDistance);
      const inkling = ambiguousHits.find((hit) => hit.slug === 'inkling');
      if (inkling) return inkling.slug;

      ambiguousHits.sort((a, b) => a.score - b.score || a.distance - b.distance);
      return ambiguousHits[0].slug;
    };

    const setChartHoveredModel = (model) => {
      if (chartHoveredModel === model) return;
      chartHoveredModel = model;
      updateActiveModel();
    };

    svg.addEventListener('pointermove', (event) => {
      setChartHoveredModel(hitTestModel(event));
    });
    svg.addEventListener('pointerleave', () => setChartHoveredModel(''));
    svg.addEventListener('pointercancel', () => setChartHoveredModel(''));

    window.addEventListener('resize', () => {
      const activeModel = getActiveModel();
      positionLegendPill(legendItems.find((item) => item.dataset.radarModel === activeModel));
    });

    radar.classList.add('is-radar-pending');
    const reveal = () => {
      radar.classList.remove('is-radar-pending');
      radar.classList.add('is-radar-visible');
    };

    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      reveal();
    } else if ('IntersectionObserver' in window) {
      const observer = new IntersectionObserver((entries) => {
        if (!entries.some((entry) => entry.isIntersecting)) return;
        observer.disconnect();
        reveal();
      }, { threshold: 0.18, rootMargin: '0px 0px -10% 0px' });
      observer.observe(radar);
    } else {
      reveal();
    }
    };

    /* Geometry construction creates more than one hundred SVG nodes and calls
       getBBox(). Defer that work until the below-the-fold chart is close to the
       viewport on both Inkling pages. */
    if ('IntersectionObserver' in window) {
      const initializeObserver = new IntersectionObserver((entries) => {
        if (!entries.some((entry) => entry.isIntersecting)) return;
        initializeObserver.disconnect();
        initialize();
      }, { threshold: 0.01, rootMargin: '200px 0px' });
      initializeObserver.observe(radar);
    } else {
      initialize();
    }
  })();
})();</script></div></div></section>
<section class="inkling-section inkling-caps-section"><h2 class="inkling-h2">The same capabilities, at every size</h2><div class="inkling-features inkling-features--caps"><article class="inkling-feature"><p class="inkling-feature-title">General intelligence</p><p class="inkling-feature-body">Performs well across knowledge, math, and science.</p></article><article class="inkling-feature"><p class="inkling-feature-title">Agentic coding &amp; tools</p><p class="inkling-feature-body">Writes code and uses tools to get work done.</p></article><article class="inkling-feature"><p class="inkling-feature-title">Audio &amp; image input</p><p class="inkling-feature-body">Natively multimodal on input, and at the open-source frontier for speech.</p></article><article class="inkling-feature"><p class="inkling-feature-title">Controllable effort</p><p class="inkling-feature-body">Adjust thinking time to balance speed and performance.</p></article><article class="inkling-feature"><p class="inkling-feature-title">Forecasting &amp; confidence</p><p class="inkling-feature-body">Makes predictions with well-calibrated confidence.</p></article><article class="inkling-feature"><p class="inkling-feature-title">Instruction following</p><p class="inkling-feature-body">Handles detailed, multi-step instructions reliably.</p></article></div></section>
<section class="inkling-section inkling-customize-section"><div class="inkling-cta-band"><div class="inkling-cta-art"><svg xmlns="http://www.w3.org/2000/svg" width="192" height="192" fill="none" aria-label="A hand-drawn region selected across a grid of Tinker nodes" role="img" viewBox="0 0 192 192"><g clip-path="url(#clip0_86_12737)"><g opacity=".3"><path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M16 6h40m20 0h40m20 0h40m-40 60h40M76 66h40M16 66h40M6 16v40m60-40v40m60 0V16m60 0v40M6 136v40M6 76v40m60-40v40m0 20v40m-50 10h40m20 0h40m0-60H76m110-50v40m-10 10h-40m0 60h40m10-50v40m-60-40v40m0-100v40m-70 10H16"/><path fill="currentColor" d="M6 12A6 6 0 1 0 6 0a6 6 0 0 0 0 12m60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12M6 72a6 6 0 1 0 0-12 6 6 0 0 0 0 12m60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m0-60a6 6 0 1 0 0-12 6 6 0 0 0 0 12m60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m0 60a6 6 0 1 0 0-12 6 6 0 0 0 0 12m0 60a6 6 0 1 0 0-12 6 6 0 0 0 0 12m-60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m-60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m-60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m0 60a6 6 0 1 0 0-12 6 6 0 0 0 0 12m60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12m60 0a6 6 0 1 0 0-12 6 6 0 0 0 0 12"/></g><mask id="path-41-inside-1_86_12737" fill="#fff"><path fill-rule="evenodd" d="M29.872 42.608a38 38 0 0 0-.548 3.145 40 40 0 0 0-.281 3.179 40 40 0 0 0-.02 3.182c.035 1.06.11 2.124.23 3.178a38 38 0 0 0 .493 3.155c.21 1.04.462 2.082.768 3.104a37 37 0 0 0 1.05 3.033 41 41 0 0 0 1.301 2.945 45 45 0 0 0 3.265 5.6 50 50 0 0 0 1.938 2.627 68 68 0 0 0 2.085 2.53 80 80 0 0 0 2.2 2.445 105 105 0 0 0 2.29 2.371q1.163 1.169 2.356 2.315 1.191 1.147 2.406 2.272 1.206 1.128 2.436 2.24 1.228 1.115 2.47 2.208a153 153 0 0 0 5.094 4.27q1.302 1.034 2.627 2.049 1.327 1.007 2.677 1.993 1.35.979 2.722 1.936c.91.644 1.828 1.274 2.737 1.914.914.644 1.829 1.278 2.723 1.941s1.783 1.33 2.636 2.035a43 43 0 0 1 2.487 2.198c.788.765 1.562 1.554 2.265 2.389a25 25 0 0 1 1.969 2.604 21 21 0 0 1 1.562 2.838c.432.975.808 1.993 1.05 3.024a14.6 14.6 0 0 1 .406 3.159 13.7 13.7 0 0 1-.321 3.164 14 14 0 0 1-1.065 3.02 15.2 15.2 0 0 1-1.763 2.73 17.5 17.5 0 0 1-2.3 2.352 21 21 0 0 1-2.702 1.951 25.5 25.5 0 0 1-2.989 1.563c-1.024.467-2.094.85-3.164 1.218-1.07.369-2.165.686-3.26.999-1.09.308-2.195.588-3.295.877-1.1.294-2.205.565-3.3.873-1.094.303-2.19.607-3.27.957-1.074.354-2.154.718-3.194 1.152-1.044.43-2.07.906-3.059 1.438a30 30 0 0 0-2.873 1.736 27 27 0 0 0-4.952 4.382 25 25 0 0 0-2.004 2.585 23 23 0 0 0-1.587 2.824 22 22 0 0 0-1.165 2.991 20 20 0 0 0-.693 3.118 19 19 0 0 0-.166 3.178c.035 1.059.15 2.128.372 3.164.215 1.036.527 2.077.944 3.057a16 16 0 0 0 1.557 2.832 17.5 17.5 0 0 0 2.124 2.497c.784.766 1.658 1.47 2.567 2.105a23 23 0 0 0 2.903 1.689c1.01.495 2.07.915 3.144 1.27 1.075.35 2.185.63 3.305.84 1.115.21 2.26.345 3.395.42a31 31 0 0 0 3.426.037 35 35 0 0 0 3.415-.313c1.13-.149 2.25-.359 3.365-.588a81 81 0 0 0 3.326-.788c1.1-.285 2.19-.607 3.284-.92 1.09-.312 2.175-.648 3.265-.97 1.09-.318 2.18-.644 3.275-.948a80 80 0 0 1 3.305-.844 58 58 0 0 1 3.35-.682 46 46 0 0 1 3.39-.452 40 40 0 0 1 3.426-.178c1.14-.014 2.285.019 3.425.103a43 43 0 0 1 3.4.392c1.131.168 2.251.392 3.366.63 1.11.243 2.215.518 3.315.812 1.1.289 2.189.611 3.279.933 1.085.322 2.17.663 3.255.999 1.085.331 2.165.677 3.255 1.003 1.085.327 2.175.654 3.27.957 1.094.303 2.194.597 3.299.868a75 75 0 0 0 6.711 1.316c1.125.168 2.26.308 3.395.434s2.275.224 3.41.313a139 139 0 0 0 6.851.345c1.14.033 2.286.056 3.426.065q1.719.016 3.43-.004c1.145-.019 2.286-.047 3.431-.094 1.14-.051 2.28-.107 3.42-.196a74 74 0 0 0 3.411-.331 55 55 0 0 0 3.385-.513 43 43 0 0 0 3.33-.756c1.095-.299 2.185-.635 3.234-1.055a26 26 0 0 0 3.049-1.451 24 24 0 0 0 2.793-1.844 21.4 21.4 0 0 0 4.505-4.769 21.4 21.4 0 0 0 1.622-2.805c.473-.961.864-1.969 1.211-2.977.341-1.013.618-2.049.859-3.085s.427-2.086.587-3.141c.161-1.05.282-2.105.382-3.164.106-1.055.176-2.119.236-3.178.055-1.06.091-2.124.121-3.183.025-1.064.035-2.128.045-3.187.01-1.064.005-2.124.01-3.188V69.414c-.005-1.064 0-2.123-.005-3.187v-6.375c0-1.06.005-2.123 0-3.183 0-1.064.005-2.128 0-3.187a226 226 0 0 0-.03-3.188 130 130 0 0 0-.091-3.182 84 84 0 0 0-.19-3.183 68 68 0 0 0-.322-3.174 48 48 0 0 0-.512-3.15 36 36 0 0 0-.764-3.103 26 26 0 0 0-1.09-3.02c-.432-.98-.919-1.95-1.496-2.865a20 20 0 0 0-1.969-2.604 19.2 19.2 0 0 0-5.214-4.093 19.7 19.7 0 0 0-3.119-1.311 22 22 0 0 0-3.32-.78 25 25 0 0 0-3.41-.303c-1.14-.028-2.29.033-3.426.13a40 40 0 0 0-3.39.467c-1.12.192-2.235.444-3.34.705-1.11.266-2.205.57-3.3.873-1.095.308-2.18.639-3.275.952-1.09.317-2.179.644-3.274.942a79 79 0 0 1-3.305.845c-1.11.252-2.23.49-3.355.672a41 41 0 0 1-3.396.425c-1.135.093-2.285.14-3.425.126a37 37 0 0 1-3.421-.2 35.7 35.7 0 0 1-6.675-1.42 41 41 0 0 1-3.199-1.138 47 47 0 0 1-3.084-1.391 56 56 0 0 1-2.979-1.577c-.979-.546-1.938-1.134-2.893-1.713-.954-.584-1.898-1.186-2.848-1.778-.949-.588-1.893-1.19-2.847-1.774-.96-.578-1.914-1.162-2.893-1.712a74 74 0 0 0-2.969-1.596 55 55 0 0 0-3.064-1.428 49 49 0 0 0-3.154-1.246 44.5 44.5 0 0 0-6.565-1.825 45 45 0 0 0-3.38-.528 46 46 0 0 0-3.415-.28 49 49 0 0 0-6.851.145 46 46 0 0 0-6.741 1.129 48 48 0 0 0-6.454 2.142 49 49 0 0 0-6.032 3.02 50 50 0 0 0-2.822 1.806 61 61 0 0 0-2.698 1.97 67 67 0 0 0-2.571 2.109 63 63 0 0 0-2.441 2.235 59 59 0 0 0-2.31 2.357 55 55 0 0 0-2.16 2.469 49 49 0 0 0-1.994 2.595 43 43 0 0 0-1.804 2.711 41 41 0 0 0-1.572 2.828 41 41 0 0 0-1.35 2.926 39 39 0 0 0-1.1 3.02 39 39 0 0 0-.825 3.089" clip-rule="evenodd"/></mask><path stroke="currentColor" stroke-width="8" d="M29.872 42.608a38 38 0 0 0-.548 3.145 40 40 0 0 0-.281 3.179 40 40 0 0 0-.02 3.182c.035 1.06.11 2.124.23 3.178a38 38 0 0 0 .493 3.155c.21 1.04.462 2.082.768 3.104a37 37 0 0 0 1.05 3.033 41 41 0 0 0 1.301 2.945 45 45 0 0 0 3.265 5.6 50 50 0 0 0 1.938 2.627 68 68 0 0 0 2.085 2.53 80 80 0 0 0 2.2 2.445 105 105 0 0 0 2.29 2.371q1.163 1.169 2.356 2.315 1.191 1.147 2.406 2.272 1.206 1.128 2.436 2.24 1.228 1.115 2.47 2.208a153 153 0 0 0 5.094 4.27q1.302 1.034 2.627 2.049 1.327 1.007 2.677 1.993 1.35.979 2.722 1.936c.91.644 1.828 1.274 2.737 1.914.914.644 1.829 1.278 2.723 1.941s1.783 1.33 2.636 2.035a43 43 0 0 1 2.487 2.198c.788.765 1.562 1.554 2.265 2.389a25 25 0 0 1 1.969 2.604 21 21 0 0 1 1.562 2.838c.432.975.808 1.993 1.05 3.024a14.6 14.6 0 0 1 .406 3.159 13.7 13.7 0 0 1-.321 3.164 14 14 0 0 1-1.065 3.02 15.2 15.2 0 0 1-1.763 2.73 17.5 17.5 0 0 1-2.3 2.352 21 21 0 0 1-2.702 1.951 25.5 25.5 0 0 1-2.989 1.563c-1.024.467-2.094.85-3.164 1.218-1.07.369-2.165.686-3.26.999-1.09.308-2.195.588-3.295.877-1.1.294-2.205.565-3.3.873-1.094.303-2.19.607-3.27.957-1.074.354-2.154.718-3.194 1.152-1.044.43-2.07.906-3.059 1.438a30 30 0 0 0-2.873 1.736 27 27 0 0 0-4.952 4.382 25 25 0 0 0-2.004 2.585 23 23 0 0 0-1.587 2.824 22 22 0 0 0-1.165 2.991 20 20 0 0 0-.693 3.118 19 19 0 0 0-.166 3.178c.035 1.059.15 2.128.372 3.164.215 1.036.527 2.077.944 3.057a16 16 0 0 0 1.557 2.832 17.5 17.5 0 0 0 2.124 2.497c.784.766 1.658 1.47 2.567 2.105a23 23 0 0 0 2.903 1.689c1.01.495 2.07.915 3.144 1.27 1.075.35 2.185.63 3.305.84 1.115.21 2.26.345 3.395.42a31 31 0 0 0 3.426.037 35 35 0 0 0 3.415-.313c1.13-.149 2.25-.359 3.365-.588a81 81 0 0 0 3.326-.788c1.1-.285 2.19-.607 3.284-.92 1.09-.312 2.175-.648 3.265-.97 1.09-.318 2.18-.644 3.275-.948a80 80 0 0 1 3.305-.844 58 58 0 0 1 3.35-.682 46 46 0 0 1 3.39-.452 40 40 0 0 1 3.426-.178c1.14-.014 2.285.019 3.425.103a43 43 0 0 1 3.4.392c1.131.168 2.251.392 3.366.63 1.11.243 2.215.518 3.315.812 1.1.289 2.189.611 3.279.933 1.085.322 2.17.663 3.255.999 1.085.331 2.165.677 3.255 1.003 1.085.327 2.175.654 3.27.957 1.094.303 2.194.597 3.299.868a75 75 0 0 0 6.711 1.316c1.125.168 2.26.308 3.395.434s2.275.224 3.41.313a139 139 0 0 0 6.851.345c1.14.033 2.286.056 3.426.065q1.719.016 3.43-.004c1.145-.019 2.286-.047 3.431-.094 1.14-.051 2.28-.107 3.42-.196a74 74 0 0 0 3.411-.331 55 55 0 0 0 3.385-.513 43 43 0 0 0 3.33-.756c1.095-.299 2.185-.635 3.234-1.055a26 26 0 0 0 3.049-1.451 24 24 0 0 0 2.793-1.844 21.4 21.4 0 0 0 4.505-4.769 21.4 21.4 0 0 0 1.622-2.805c.473-.961.864-1.969 1.211-2.977.341-1.013.618-2.049.859-3.085s.427-2.086.587-3.141c.161-1.05.282-2.105.382-3.164.106-1.055.176-2.119.236-3.178.055-1.06.091-2.124.121-3.183.025-1.064.035-2.128.045-3.187.01-1.064.005-2.124.01-3.188V69.414c-.005-1.064 0-2.123-.005-3.187v-6.375c0-1.06.005-2.123 0-3.183 0-1.064.005-2.128 0-3.187a226 226 0 0 0-.03-3.188 130 130 0 0 0-.091-3.182 84 84 0 0 0-.19-3.183 68 68 0 0 0-.322-3.174 48 48 0 0 0-.512-3.15 36 36 0 0 0-.764-3.103 26 26 0 0 0-1.09-3.02c-.432-.98-.919-1.95-1.496-2.865a20 20 0 0 0-1.969-2.604 19.2 19.2 0 0 0-5.214-4.093 19.7 19.7 0 0 0-3.119-1.311 22 22 0 0 0-3.32-.78 25 25 0 0 0-3.41-.303c-1.14-.028-2.29.033-3.426.13a40 40 0 0 0-3.39.467c-1.12.192-2.235.444-3.34.705-1.11.266-2.205.57-3.3.873-1.095.308-2.18.639-3.275.952-1.09.317-2.179.644-3.274.942a79 79 0 0 1-3.305.845c-1.11.252-2.23.49-3.355.672a41 41 0 0 1-3.396.425c-1.135.093-2.285.14-3.425.126a37 37 0 0 1-3.421-.2 35.7 35.7 0 0 1-6.675-1.42 41 41 0 0 1-3.199-1.138 47 47 0 0 1-3.084-1.391 56 56 0 0 1-2.979-1.577c-.979-.546-1.938-1.134-2.893-1.713-.954-.584-1.898-1.186-2.848-1.778-.949-.588-1.893-1.19-2.847-1.774-.96-.578-1.914-1.162-2.893-1.712a74 74 0 0 0-2.969-1.596 55 55 0 0 0-3.064-1.428 49 49 0 0 0-3.154-1.246 44.5 44.5 0 0 0-6.565-1.825 45 45 0 0 0-3.38-.528 46 46 0 0 0-3.415-.28 49 49 0 0 0-6.851.145 46 46 0 0 0-6.741 1.129 48 48 0 0 0-6.454 2.142 49 49 0 0 0-6.032 3.02 50 50 0 0 0-2.822 1.806 61 61 0 0 0-2.698 1.97 67 67 0 0 0-2.571 2.109 63 63 0 0 0-2.441 2.235 59 59 0 0 0-2.31 2.357 55 55 0 0 0-2.16 2.469 49 49 0 0 0-1.994 2.595 43 43 0 0 0-1.804 2.711 41 41 0 0 0-1.572 2.828 41 41 0 0 0-1.35 2.926 39 39 0 0 0-1.1 3.02 39 39 0 0 0-.825 3.089Z" clip-rule="evenodd" mask="url(#path-41-inside-1_86_12737)"/></g><defs><clipPath id="clip0_86_12737"><path fill="#fff" d="M0 0h192v192H0z"/></clipPath></defs></svg></div><h2 class="inkling-h2">Ready to customize</h2><p class="inkling-cta-band-body">Grab the open weights from Hugging Face, or fine-tune Inkling for your domain on Tinker, our training platform.</p><div class="inkling-cta-actions"><a class="btn primary" href="https://tinker.thinkingmachines.ai/playground?utm_source=blog&amp;utm_campaign=inkling_model_release" target="_blank" rel="noopener" data-astro-cid-ekguhzzh="true">Try on Tinker</a><a class="btn plain btn-with-chevron" href="https://huggingface.co/collections/thinkingmachines/inkling" target="_blank" rel="noopener" data-astro-cid-ekguhzzh="true">Hugging Face<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" data-astro-cid-ekguhzzh><path fill-rule="evenodd" clip-rule="evenodd" d="M6.56701 3.34347C6.78189 3.14201 7.1194 3.1529 7.32086 3.36779L11.3208 7.63446C11.5131 7.8396 11.5131 8.15884 11.3208 8.36399L7.32086 12.6306C7.1194 12.8456 6.78189 12.8565 6.56701 12.655C6.35212 12.4535 6.34123 12.116 6.54269 11.9011L10.2007 7.99922L6.54269 4.09733C6.34123 3.88244 6.35212 3.54492 6.56701 3.34347Z" fill="currentColor" data-astro-cid-ekguhzzh></path></svg></a></div></div></section>
<p class="inkling-rightsholder-notice">Rightsholders and their authorized representatives can contact us about these models at <a href="/cdn-cgi/l/email-protection#e08c8587818ca09488898e8b898e878d818388898e8593ce8189"><span class="__cf_email__" data-cfemail="fb979e9c9a97bb8f9392959092959c969a989392959e88d59a92">[email&#160;protected]</span></a>.</p></article></main><script data-cfasync="false" src="/cdn-cgi/scripts/5c5dd728/cloudflare-static/email-decode.min.js"></script><script type="module">
  import mediumZoom from 'https://cdnjs.cloudflare.com/ajax/libs/medium-zoom/1.1.0/medium-zoom.esm.min.js';

  const getZoomContainer = () => ({
    width: window.innerWidth,
    height: window.innerHeight,
    top: window.innerHeight * 0.15,
    right: window.innerWidth * 0.15,
    bottom: window.innerHeight * 0.15,
    left: window.innerWidth * 0.15,
  });

  const zoom = mediumZoom('[data-zoomable]', {
    background: '#fff',
    container: getZoomContainer(),
    margin: 0,
  });

  window.addEventListener('resize', () => {
    zoom.update({ container: getZoomContainer() });
  });
</script><footer id="footer"><div class="footer-columns"><div class="footer-links"><div class="footer-column"><h4>Models</h4><ul><li><a href="/news/introducing-inkling/">Inkling</a></li><li><a href="/news/inkling-small/">Inkling-Small</a></li><li><a href="/blog/interaction-models/">Interaction</a></li></ul></div><div class="footer-column"><h4>Tinker</h4><ul><li><a href="/tinker/">Overview</a></li><li><a href="https://tinker-docs.thinkingmachines.ai/tinker/">Docs</a></li></ul></div><div class="footer-column"><h4>Company</h4><ul><li><a href="/blog/the-future-worth-building-is-human/">About</a></li><li><a href="/blog/">Connectionism</a></li><li><a href="/news/">News</a></li><li><a href="https://jobs.ashbyhq.com/ThinkingMachines">Careers</a></li></ul></div><div class="footer-column"><h4>Legal</h4><ul><li><a href="/legal/terms/">Terms of service</a></li><li><a href="/legal/privacy/">Privacy notice</a></li></ul><h4 class="footer-subheading">Security</h4><ul><li><a href="/security/disclosure-policy/">Disclosure policy</a></li></ul></div><div class="footer-column"><h4>Connect</h4><ul><li><a href="https://x.com/thinkymachines">X</a></li><li><a href="https://www.linkedin.com/company/thinkingmachinesai/">LinkedIn</a></li><li><a href="https://www.youtube.com/@thinkymachines">YouTube</a></li></ul></div></div></div><div class="footer-copyright"><span>Thinking Machines Lab © 2026</span></div></footer></div><script src="/js/script.min.79c2d79958493c8b72e12b13fa5d89ef49e3ae95a7aea5b82dbcc7e382c01b5a.js" integrity="sha256-ecLXmVhJPIty4SsT+l2J70njrpWnrqW4LbzH44LAG1o=" defer></script><script type="module" src="https://static.cloudflareinsights.com/beacon.min.js/v31edd6df95cf4e85bb4c19e7a9bdbcba1788362987495" integrity="sha512-iIg7k2xntmwu6/uSb5tpc/hySgZc4eoL31yB29W6tJFo2akwjPWcEqnCEdJvGexCL0KEQwVYv5BlowfhVz26hg==" data-cf-beacon='{"version":"2024.11.0","token":"cda777236064450ebb47f38aeb8df789","spa":2}' crossorigin="anonymous"></script>
</body></html>