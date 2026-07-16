import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

type WikiSection = {
  id: string
  title: string
  summary: string
  body: string[]
  formula?: string
}

const zhSections: WikiSection[] = [
  {
    id: 'first-day',
    title: '新手：第一天怎么做',
    summary: '把一次小而完整的生意做完，再扩张。',
    body: [
      '1. 建设并放置农场或其他生产建筑；点击建筑，选定一种商品和数量后开始生产。',
      '2. 生产完成后领取货物。货物会进入仓库；加工商品会同时消耗配方所需的原料。',
      '3. 将商品上架到零售建筑，或在市场挂卖单。先做小批量交易，确认现金与仓储空间都健康。',
      '4. 下线也没关系：生产按服务器时间继续；回来后领取、补货、重新排产。',
    ],
  },
  {
    id: 'daily-loop',
    title: '核心循环：上线、生产、卖出、下线',
    summary: '每次上线优先处理“已经完成”的事。',
    body: [
      '检查生产队列：领取可领取的货物，再为闲置建筑安排下一单。',
      '检查仓库：原料不足时去市场采购；成品积压时上架零售或下卖单。',
      '检查现金和挂单：市场成交和零售收入会改变下一轮可购买的原料与建筑。',
      '离线前把生产线和货架填好；离线后再回来结算结果。这是新港最稳定的休闲节奏。',
    ],
  },
  {
    id: 'production',
    title: '生产线规则',
    summary: '一座建筑 = 一条生产线 = 同时一种商品。',
    body: [
      '一座建筑同一时间只能有一个未完成的生产订单。完成后请领取全部产出，或取消订单，才能开始下一种商品。',
      'Q0 原材料不需要投入；Q0 加工品和成品会显示基础配方，并在开始生产前从仓库预留所需原料。因此不要把同一批原料同时许诺给市场和工厂。',
      '建筑等级会提高产出速度。生产期间也按这栋建筑固定员工数 × $345/小时累积工资；在领取或取消订单时，系统会按实际运行秒数一次结算。每个订单最多 48 小时，建议用较短批次观察需求再加量。',
    ],
    formula: '当前生产时长 = max(30 秒, ceil(数量 ÷ (基础每小时产量 × 建筑等级 × 生产加成) × 3600))',
  },
  {
    id: 'quality',
    title: '品质链：Q0 到 Q12',
    summary: '更高品质来自上一品质，不是凭空选择出来的标签。',
    body: [
      '同一种商品的 Q0–Q12 是 13 个互相独立的库存栈。生产、领取、取消退款与零售都会保留品质；高品质库存不能冒充低品质原料，低品质也不能自动顶替高品质。',
      '每种商品初始只能生产 Q0。请在“品质实验室”花费现金逐级解锁 Q1–Q12；不能跳级。研究只授予该公司、该商品的生产许可，不赠送库存。重复提交同一个目标品质不会重复扣钱。',
      '生产 Q0 加工品时使用普通配方和 Q0 原料。生产 Q1–Q12 加工品时，使用前一品质的配方原料，且每项投入是普通配方的 2 倍。例如普通配方需要 2 面粉，生产 Q4 时就需要 4 个 Q3 面粉。',
      '原材料没有普通配方，因此升级品质采用精炼：每生产 1 个 Qn 原材料，需要 2 个同商品的 Q(n-1) 原材料。所有投入在开始生产时原子预留，库存不足不会启动半个订单。',
      '每级品质只在零售端提供 +2% 需求速度，Q12 共 +24%。品质不会抬高成交单价，也不会取消高价销售惩罚；离谱定价依旧可能让工资超过收入。当前交易所仍是 Q0 现货市场。',
    ],
    formula: '研究 Qn = 向上取整到 $10（$1,000 × 2^(商品阶级-1) × 1.2^(n-1)）；Qn 投入 = 2 × 基础配方的 Q(n-1) 原料；零售速度倍率 = 1 + 0.02 × n',
  },
  {
    id: 'upgrades',
    title: '建筑升级：施工时间与停工',
    summary: '确认施工后先扣费用，计时完成才提升等级。',
    body: [
      '升级前需要先领取或取消当前生产订单；零售建筑也要等正在销售的批次售罄。升级会立即预留现金，但建筑等级和产能会在施工完成后才提升。施工期间不能生产、开始新的销售、移动、收纳或拆除该建筑。',
      '每种建筑有自己的基础施工分钟数。当前等级为 1、2、3、4、5 时，施工时间分别为基础时间的 1×、2×、3×、6×、10×。',
      '从 6 级开始继续使用三角数倍率：15×、21×、28×……。离线不会暂停计时；再次打开建筑时会完成已到期的施工。',
    ],
    formula: '施工时间 = 建筑基础分钟 × [1, 2, 3, 6, 10, 15, 21, …]（升级完成后才提高一级）',
  },
  {
    id: 'retail',
    title: '零售：把货架变成现金',
    summary: '选择数量与价格，启动一个不可撤销的销售批次。',
    body: [
      '零售建筑只会卖它允许经营的商品。开始前选择商品、数量和销售价；推荐价由商品当前市场来源成本、这栋建筑的固定员工工资，以及实时消费者需求共同计算。',
      '点击“开始销售”后，这批货物和价格会锁定，不能撤货、追加或改价。商品全部售罄后，才能用新价格开始下一批销售。',
      '每种商品都有自己的基础需求，需求会随时间和玩家买卖盘在 55%–145% 间变化。价格高于推荐价会按价格比的平方减速；极高价格几乎不成交，但只要货架仍在销售，员工工资仍持续发生。低流量商品的零星需求会被保留到下一次结算，不会因为页面轮询而丢失。',
      '不同零售建筑的货架独立结算；同一商品也不会从另一座店铺误扣库存。',
    ],
  },
  {
    id: 'market',
    title: '市场：采购与出售',
    summary: '用限价单保护你的现金和利润。',
    body: [
      '买单设定你愿意支付的最高价，卖单设定你愿意接受的最低价。成交前请确认数量、价格和仓库空间。',
      '市场首次打开时会为全部可交易商品检查 Bot 流动性并补齐买卖盘。移动参考价会把上游配方成本、生产建筑的工资与目标利润传递到下游，并根据玩家买卖盘失衡变化，不再使用固定的 $10。',
      '价格必须符合交易所最小变动单位（tick）。例如价格在 5–19.999 时，最小变动为 0.1。',
      '交易手续费按成交额的 4% 向上取整。计算利润时，也要加上原料与运输成本。',
    ],
    formula: '手续费 = ceil(数量 × 成交价 × 0.04)',
  },
  {
    id: 'economy',
    title: '经济与建筑底层规则',
    summary: '升级提升能力，仓库限制周转，现金决定扩张速度。',
    body: [
      '建筑采购价优先按市场材料价计算；缺少有效市场价时使用成本单位 × 3450 的兜底价。',
      '单建筑每 tick 的基础工资为 345 × 工资系数 × 建筑规模。工资、研究和行政能力影响经营效率；商品品质目前只提高零售需求速度。',
      '仓库满时，领取、采购和补货会受阻。保留余量比把库存堆满更安全。',
    ],
    formula: '市场采购价 = Σ(材料市价 × 成本单位 × qp)；兜底价 = 成本单位 × 3450',
  },
  {
    id: 'executives',
    title: '高管：四个岗位、四项技能',
    summary: '高管不是隐藏的等级光环；你要看技能，再把人放到正确的岗位。',
    body: [
      '高管市场每小时刷新。候选人有专长岗位（COO、CFO、CMO、CTO）和管理、会计、沟通、科学四项技能；专长会让对应技能更高。招聘扣除页面标明的一次性费用，同一候选人不能被同一公司重复招聘。',
      '每个领导岗位同一时间只能有一位负责人。改派某个岗位会自动让原负责人变为未分配；未分配的高管不会给当前生产或零售提供隐藏加成。',
      'CTO 的科学技能已接入生产：有效科学每点提供 +2% 生产速度，最高为 3 倍速度。CMO 的有效沟通每 2 点提供 +1% 零售需求速度，最高 +50%。高价格仍会减慢销售，CMO 不会让离谱价格正常成交。',
      'COO 的管理与 CFO 的会计会保留到公司档案中，供后续行政开销和财务工具使用。当前版本没有行政费用或手续费折扣，因此它们不会悄悄改变工资或固定 4% 市场手续费。',
      '培养会立即扣除明确显示的现金费用：所有技能 +1，专长技能再 +3。系统不会显示一个后端已经完成的假培训计时。',
    ],
    formula: '有效技能：x≤60→x；60<x≤80→60+(x-60)/2；x>80→70+(x-80)/2。CTO 速度倍率 = (100 + 2×有效科学)/100',
  },
]

const enSections: WikiSection[] = [
  {
    id: 'first-day', title: 'New player: your first day', summary: 'Finish one small business loop before expanding.',
    body: ['1. Build and place a farm or another producer; select a product and quantity, then start a run.', '2. Collect completed goods into the warehouse. Processed goods reserve their recipe inputs when the run starts; the building’s fixed payroll also accrues for its actual run time and settles when you claim or cancel.', '3. Stock a retail shelf or place a market sell order. Start with small batches and keep an eye on cash and capacity.', '4. Production continues on server time while you are away. Return to collect, settle payroll, restock, and schedule the next run.'],
  },
  {
    id: 'daily-loop', title: 'Core loop: log in, make, sell, log out', summary: 'Handle completed work first every time you return.',
    body: ['Check production: collect what is ready, then schedule idle buildings.', 'Check the warehouse: buy inputs when short; stock shelves or sell finished goods when full.', 'Check cash and orders: market fills and retail sales shape what you can afford next.', 'Before leaving, fill production lines and shelves. Come back to settle results.'],
  },
  {
    id: 'production', title: 'Production line rules', summary: 'One building = one line = one product at a time.',
    body: ['A building can have only one unfinished run. Collect all output or cancel the run before starting another product.', 'Q0 raw goods need no inputs. Q0 processed and finished goods show their base recipe and reserve those ingredients from warehouse inventory when a run starts.', 'Production also accrues the building’s fixed worker payroll (workers × $345/hour) for its actual running time. It settles exactly once when you claim output or cancel the run.', 'Building levels improve output speed. Runs are capped at 48 hours; shorter batches are safer while learning demand.'],
    formula: 'Current duration = max(30s, ceil(quantity ÷ (base hourly output × building level × production modifier) × 3600))',
  },
  {
    id: 'quality', title: 'Quality chain: Q0 to Q12', summary: 'Higher quality is made from the previous tier; it is not a free label.',
    body: ['Q0 through Q12 of the same resource are 13 separate warehouse stacks. Production, collection, cancellation refunds, and retail preserve quality. Tiers never substitute for one another.', 'Each product begins with Q0 production only. Spend cash in the Quality Lab to unlock Q1 through Q12 in order; levels cannot be skipped. Research grants only this company and product a production licence, not free inventory. Replaying the same target quality never charges twice.', 'A Q0 processed good uses its normal recipe at Q0. A Q1–Q12 processed good consumes the previous quality of every recipe ingredient at twice the normal quantity. If a base recipe needs 2 Flour, a Q4 run needs 4 Q3 Flour per output.', 'Raw resources have no base recipe, so they are refined: each Qn raw unit consumes 2 units of the same resource at Q(n-1). All inputs are atomically reserved when production starts; a shortage cannot create a partially funded run.', 'Each quality tier adds 2% retail demand speed, up to +24% at Q12. Quality does not raise the sale price or bypass the high-price penalty, so a wildly overpriced batch can still lose money to payroll. The exchange currently remains a Q0 spot market.'],
    formula: 'Research Qn = ceil to $10 ($1,000 × 2^(product tier-1) × 1.2^(n-1)); Qn inputs = 2 × base recipe at Q(n-1); retail speed = 1 + 0.02 × n',
  },
  {
    id: 'upgrades', title: 'Building upgrades: construction time and downtime', summary: 'Construction reserves its cost now; the level increases only at completion.',
    body: ['Collect or cancel an active production run before upgrading; retail buildings must also wait for active sale batches to sell out. Starting an upgrade reserves cash immediately, but the level and output increase only when construction completes. The building cannot produce, start a new sale, move, stash, or be demolished while it is under construction.', 'Every building family has its own base construction minutes. At current levels 1, 2, 3, 4, and 5, upgrades take 1×, 2×, 3×, 6×, and 10× that base time.', 'From level 6 the triangular sequence continues: 15×, 21×, 28×, and so on. The timer continues while you are offline and completes the next time the building is checked.'],
    formula: 'construction time = building base minutes × [1, 2, 3, 6, 10, 15, 21, …]; the level increases on completion',
  },
  {
    id: 'retail', title: 'Retail: turn shelves into cash', summary: 'Choose quantity and price, then commit an irreversible sale batch.',
    body: ['Retail buildings sell only allowed goods. Before starting, choose the product, quantity, and sale price. The recommendation combines the item’s current market source cost, this building’s fixed payroll, and live customer demand.', 'After you press Start selling, that batch and price are locked: no unstocking, topping up, or repricing. Start a new batch after every unit sells.', 'Every item has its own base demand. Time and public player orders move that demand between 55% and 145%. Above the recommendation, demand falls with the square of the price ratio; an extreme price can leave a batch effectively unsold.', 'Every building has a fixed worker count from its type and level. While a retail batch is active, its payroll is workers × $345 per hour; idle or upgrading buildings do not charge this payroll. A slow, expensive batch can therefore lose money to wages even while it waits for buyers.', 'Shelves settle independently by building, even when two shops sell the same resource.'],
  },
  {
    id: 'market', title: 'Market: buy and sell', summary: 'Use limit orders to protect cash and margin.',
    body: ['A buy order sets your maximum price; a sell order sets your minimum. Check quantity, price, and warehouse space before confirming.', 'On first open, the market checks every tradable resource and fills missing Bot liquidity on both sides. Its moving reference carries upstream recipe costs forward, includes the producer’s wage-and-profit baseline, and responds to player order imbalance instead of using a fixed $10.', 'Prices must respect the exchange tick size. For example, the tick between 5 and 19.999 is 0.1.', 'The exchange takes a 4% fee rounded up. Include materials and transport when estimating profit.'],
    formula: 'fee = ceil(quantity × filled price × 0.04)',
  },
  {
    id: 'economy', title: 'Economy and building rules', summary: 'Upgrades add capacity, warehouses constrain flow, and cash controls expansion.',
    body: ['Building purchase cost uses market-priced materials when available; otherwise it falls back to cost units × 3450.', 'Base wages per building tick are 345 × salary modifier × building size. Wages, research, and administration feed operating efficiency; product quality currently affects retail demand speed only.', 'A full warehouse blocks collection, buying, and restocking. Keep headroom instead of filling it completely.'],
    formula: 'market purchase cost = Σ(material price × cost units × qp); fallback = cost units × 3450',
  },
  {
    id: 'executives', title: 'Executives: four positions, four skills', summary: 'Executives are not a hidden level aura: inspect skills, then put the right person in the right chair.',
    body: ['The executive market refreshes hourly. Every candidate has a specialty (COO, CFO, CMO, or CTO) plus Management, Accounting, Communication, and Science skills. Their specialty is weighted toward its matching skill. Recruitment deducts the shown one-time cost and the same company cannot hire the same candidate twice.', 'Each leadership chair has one holder. Moving a new executive into a chair unassigns the previous holder. Unassigned executives do not provide hidden production or retail buffs.', 'An assigned CTO turns effective Science into +2% production speed per point, capped at a 3× speed multiplier. An assigned CMO turns every 2 effective Communication points into +1% retail demand speed, capped at +50%. A CMO cannot make an extreme price sell normally: the price penalty still applies.', 'COO Management and CFO Accounting are retained in the company record for upcoming administration and finance tools. The minimal economy does not yet charge administration overhead or discount the published 4% exchange fee, so these roles do not secretly change wages or completed trades.', 'Development is immediate and server-charged: every skill gains +1 and the specialty gains another +3. The interface never shows a training timer for a change the server has already completed.'],
    formula: 'effective skill: x≤60→x; 60<x≤80→60+(x-60)/2; x>80→70+(x-80)/2. CTO multiplier = (100 + 2 × effective Science) / 100',
  },
]

export function GameWikiPage() {
  const { i18n } = useTranslation()
  const sections = i18n.language.toLowerCase().startsWith('zh') ? zhSections : enSections
  const [selectedID, setSelectedID] = useState(sections[0].id)
  const selected = useMemo(() => sections.find((section) => section.id === selectedID) ?? sections[0], [sections, selectedID])

  return (
    <main className="min-h-full bg-[#f6ecd6] px-4 py-5 text-amber-950 sm:px-7 sm:py-7">
      <div className="mx-auto grid max-w-5xl gap-4 lg:grid-cols-[210px_minmax(0,1fr)]">
        <aside className="rounded-xl border border-amber-300/70 bg-[#fffaf0] p-3 shadow-sm lg:sticky lg:top-4 lg:h-fit">
          <div className="mb-3 px-2 text-[10px] font-black uppercase tracking-[0.2em] text-amber-600">New Haven Wiki</div>
          <div className="flex gap-1 overflow-x-auto lg:flex-col">
            {sections.map((section) => (
              <button
                key={section.id}
                onClick={() => setSelectedID(section.id)}
                className={`shrink-0 rounded-lg px-3 py-2 text-left text-xs font-bold transition-colors lg:shrink ${selected.id === section.id ? 'bg-amber-800 text-amber-50 shadow-sm' : 'text-amber-800 hover:bg-amber-100'}`}
              >
                {section.title}
              </button>
            ))}
          </div>
        </aside>

        <article className="rounded-2xl border-2 border-amber-300/70 bg-[#fffaf0] p-5 shadow-md shadow-amber-950/10 sm:p-7">
          <div className="mb-1 text-[10px] font-black uppercase tracking-[0.24em] text-amber-600">Player handbook · current playable rules</div>
          <h1 className="text-2xl font-black tracking-tight text-amber-950 sm:text-3xl">{selected.title}</h1>
          <p className="mt-2 border-l-4 border-amber-500 pl-3 text-sm font-semibold leading-relaxed text-amber-800">{selected.summary}</p>

          <ol className="mt-6 space-y-3 text-sm leading-6 text-amber-900">
            {selected.body.map((paragraph) => <li key={paragraph} className="rounded-lg bg-amber-50 px-3 py-2">{paragraph}</li>)}
          </ol>

          {selected.formula && (
            <div className="mt-6 rounded-xl border border-cyan-800/20 bg-cyan-50 px-4 py-3">
              <div className="text-[10px] font-black uppercase tracking-[0.18em] text-cyan-800">Rule formula</div>
              <code className="mt-1 block break-words text-xs font-bold leading-5 text-cyan-950">{selected.formula}</code>
            </div>
          )}

          <div className="mt-7 rounded-xl bg-amber-900 px-4 py-3 text-xs font-semibold leading-5 text-amber-50">
            {i18n.language.toLowerCase().startsWith('zh')
              ? '这本百科会随可玩规则更新。若界面行为与本页不一致，以本页标注的“当前可玩规则”为准，并欢迎通过游戏内聊天反馈。'
              : 'This wiki is updated with playable rules. If an interface appears to disagree, follow the “current playable rules” marked here and report it through in-game chat.'}
          </div>
        </article>
      </div>
    </main>
  )
}
