<script lang="ts">
	import '$lib/styles/grifel.css';
	import type { StatBlock } from '$lib/types';

	let {
		statblock,
		nameRu,
		nameEn,
		sourceUrl,
		imageUrl,
		notes = '',
		onNotesChange
	}: {
		statblock: StatBlock;
		nameRu: string;
		nameEn?: string;
		sourceUrl?: string;
		imageUrl?: string;
		notes?: string;
		onNotesChange?: (value: string) => void;
	} = $props();

	let showNotes = $state(false);

	const abilityLabels: [keyof StatBlock['abilities'], string][] = [
		['str', 'СИЛ'],
		['dex', 'ЛОВ'],
		['con', 'ТЕЛ'],
		['int', 'ИНТ'],
		['wis', 'МДР'],
		['cha', 'ХАР']
	];

	function fmtMod(mod: number): string {
		return mod >= 0 ? `+${mod}` : `${mod}`;
	}

	function entries(rec?: Record<string, string>): [string, string][] {
		return rec ? Object.entries(rec) : [];
	}

	const featureSections = $derived(
		[
			{ title: 'Особенности', id: 'osobennosti', features: statblock.traits },
			{ title: 'Действия', id: 'deystviya', features: statblock.actions },
			{ title: 'Бонусные действия', id: 'bonus', features: statblock.bonusActions },
			{ title: 'Реакции', id: 'reaktsii', features: statblock.reactions },
			{ title: 'Легендарные действия', id: 'legendary', features: statblock.legendaryActions }
		].filter((s) => s.features?.length)
	);

	function isRecharge(name: string): boolean {
		return /перезарядка/i.test(name);
	}

	function splitRecharge(name: string): { label: string; recharge: string | null } {
		const m = name.match(/^(.*?)\s*(перезарядка[^.]*)\.?$/i);
		if (!m) return { label: name, recharge: null };
		return { label: m[1].trim(), recharge: m[2].trim() };
	}
</script>

<div class="gr-statblock">
	<div class="gr-image" class:gr-image-photo={imageUrl}>
		{#if imageUrl}
			<img class="gr-image-img" src={imageUrl} alt={nameRu} />
		{:else}
			<span class="gr-image-caption">{nameEn ?? nameRu}</span>
		{/if}
		{#if sourceUrl}
			<a class="gr-dndsu-btn" href={sourceUrl} target="_blank" rel="noreferrer">dnd.su ↗</a>
		{/if}
	</div>

	<div class="gr-body">
		<header class="gr-header">
			<h2>{nameRu}{#if nameEn}<span class="gr-name-en"> {nameEn}</span>{/if}</h2>
			{#if statblock.sizeRu || statblock.type}
				<p class="gr-meta">
					{statblock.sizeRu}
					{statblock.type}{#if statblock.alignment}, {statblock.alignment}{/if}
				</p>
			{/if}
		</header>

		<div class="gr-top-stats">
			{#if statblock.armorClass}
				<div><span class="gr-label">Класс доспеха</span><span class="gr-value">{statblock.armorClass} {#if statblock.armorSource}<em>({statblock.armorSource})</em>{/if}</span></div>
			{/if}
			{#if statblock.hitPoints}
				<div><span class="gr-label">Хиты</span><span class="gr-value">{statblock.hitPoints}{#if statblock.hitDice}<em> ({statblock.hitDice})</em>{/if}</span></div>
			{/if}
			{#if statblock.speeds}
				<div>
					<span class="gr-label">Скорость</span>
					<span class="gr-value"
						>{Object.entries(statblock.speeds)
							.map(([k, v]) => (k === 'ходьба' ? v : `${k} ${v}`))
							.join(', ')}</span
					>
				</div>
			{/if}
		</div>

		<div class="gr-abilities">
			{#each abilityLabels as [key, label] (key)}
				<div class="gr-ability">
					<div class="gr-ability-label">{label}</div>
					<div class="gr-ability-score">{statblock.abilities[key]?.score ?? '—'}</div>
					<div class="gr-ability-mod">{fmtMod(statblock.abilities[key]?.mod ?? 0)}</div>
				</div>
			{/each}
		</div>

		<dl class="gr-details">
			{#if entries(statblock.savingThrows).length}
				<div><dt>Спасброски</dt><dd>{entries(statblock.savingThrows).map(([k, v]) => `${k} ${v}`).join(', ')}</dd></div>
			{/if}
			{#if entries(statblock.skills).length}
				<div><dt>Навыки</dt><dd>{entries(statblock.skills).map(([k, v]) => `${k} ${v}`).join(', ')}</dd></div>
			{/if}
			{#if statblock.vulnerabilities?.length}
				<div><dt>Уязвимости</dt><dd>{statblock.vulnerabilities.join(', ')}</dd></div>
			{/if}
			{#if statblock.resistances?.length}
				<div><dt>Сопротивления</dt><dd>{statblock.resistances.join(', ')}</dd></div>
			{/if}
			{#if statblock.immunities?.length}
				<div><dt>Иммунитеты</dt><dd>{statblock.immunities.join(', ')}</dd></div>
			{/if}
			{#if statblock.conditionImmunities?.length}
				<div><dt>Иммунитет к состояниям</dt><dd>{statblock.conditionImmunities.join(', ')}</dd></div>
			{/if}
			{#if statblock.senses}
				<div><dt>Чувства</dt><dd>{statblock.senses}</dd></div>
			{/if}
			{#if statblock.languages}
				<div><dt>Языки</dt><dd>{statblock.languages}</dd></div>
			{/if}
			{#if statblock.challengeRating}
				<div>
					<dt>Опасность</dt>
					<dd>
						{statblock.challengeRating}
						{#if statblock.experiencePoints}({statblock.experiencePoints} опыта){/if}
						{#if statblock.proficiencyBonus}· Бонус мастерства {fmtMod(statblock.proficiencyBonus)}{/if}
					</dd>
				</div>
			{/if}
			{#if statblock.habitat}
				<div><dt>Среда обитания</dt><dd>{statblock.habitat}</dd></div>
			{/if}
		</dl>

		{#each featureSections as section (section.id)}
			<section class="gr-features" id={section.id}>
				<h3>{section.title}</h3>
				{#each section.features ?? [] as f (f.name + f.description.slice(0, 20))}
					{@const { label, recharge } = splitRecharge(f.name)}
					{#if recharge}
						<div class="gr-action gr-action-recharge">
							<p><b>{label}</b> <span class="gr-recharge-badge">{recharge}</span></p>
							<p>{f.description}</p>
						</div>
					{:else}
						<p class="gr-action"><b>{f.name}.</b> {f.description}</p>
					{/if}
				{/each}
			</section>
		{/each}

		{#if statblock.spellcasting}
			<section class="gr-features"><h3>Заклинания</h3><p class="gr-action">{statblock.spellcasting}</p></section>
		{/if}
		{#if statblock.lairActions}
			<section class="gr-features" id="logovo"><h3>Логово</h3><p class="gr-action">{statblock.lairActions}</p></section>
		{/if}
		{#if statblock.regionalEffects}
			<section class="gr-features"><h3>Региональные эффекты</h3><p class="gr-action">{statblock.regionalEffects}</p></section>
		{/if}

		{#if statblock.sourceBook}
			<p class="gr-source">Источник: {statblock.sourceBook}</p>
		{/if}

		<div class="gr-footer-actions">
			{#if featureSections.some((s) => s.id === 'osobennosti')}
				<a class="gr-pill-btn" href="#osobennosti">Особенности</a>
			{/if}
			{#if statblock.lairActions}
				<a class="gr-pill-btn" href="#logovo">Логово</a>
			{/if}
			<button class="gr-pill-btn gr-pill-btn-filled" onclick={() => (showNotes = !showNotes)}>
				+ Добавить заметку
			</button>
		</div>

		{#if showNotes}
			<label class="gr-notes">
				Заметки
				<textarea value={notes} onchange={(e) => onNotesChange?.((e.target as HTMLTextAreaElement).value)}
				></textarea>
			</label>
		{/if}
	</div>
</div>

<style>
	.gr-statblock {
		font-family: var(--gr-font-body);
		color: var(--gr-ink-soft);
		height: 100%;
		overflow-y: auto;
	}
	.gr-image {
		position: relative;
		height: 8rem;
		background: linear-gradient(135deg, #ddc9a3, #c2ab82);
		background-image:
			repeating-linear-gradient(45deg, rgba(58, 44, 28, 0.04) 0 2px, transparent 2px 14px),
			linear-gradient(135deg, #ddc9a3, #c2ab82);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.gr-image-caption {
		font-family: ui-monospace, Menlo, monospace;
		font-size: 0.75rem;
		letter-spacing: 0.08em;
		color: #7a6a4a;
	}
	.gr-image-photo {
		background: var(--gr-ink);
	}
	.gr-image-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.gr-dndsu-btn {
		position: absolute;
		right: 0.875rem;
		top: 0.8rem;
		font-family: var(--gr-font-display);
		font-size: 0.625rem;
		font-weight: 600;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--gr-ink-soft);
		background: rgba(243, 227, 198, 0.92);
		border-radius: var(--gr-radius-md);
		padding: 0.375rem 0.7rem;
		text-decoration: none;
	}
	.gr-body {
		padding: 1.1rem 1.25rem 1.5rem;
	}
	.gr-header h2 {
		margin: 0;
		font-family: var(--gr-font-display);
		font-size: 1.4rem;
		font-weight: 700;
		color: var(--gr-ink);
	}
	.gr-name-en {
		font-weight: 400;
		font-style: italic;
		font-size: 0.95rem;
		color: var(--gr-ink-muted);
	}
	.gr-meta {
		font-style: italic;
		color: var(--gr-ink-muted);
		margin: 0.2rem 0 0;
		font-size: 0.9rem;
	}
	.gr-top-stats {
		display: flex;
		flex-wrap: wrap;
		gap: 1.5rem;
		margin: 0.9rem 0;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--gr-parchment-border);
	}
	.gr-top-stats .gr-label {
		display: block;
		font-family: var(--gr-font-display);
		font-size: 0.65rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--gr-ink-muted);
		margin-bottom: 0.2rem;
	}
	.gr-top-stats .gr-value {
		font-weight: 700;
		color: var(--gr-ink);
	}
	.gr-top-stats .gr-value em {
		font-weight: 400;
		font-style: italic;
		color: var(--gr-ink-muted);
	}
	.gr-abilities {
		display: grid;
		grid-template-columns: repeat(6, 1fr);
		gap: 0.5rem;
		margin: 0.9rem 0;
		text-align: center;
	}
	.gr-ability {
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		border-radius: var(--gr-radius-lg);
		padding: 0.5rem 0.25rem;
	}
	.gr-ability-label {
		font-family: var(--gr-font-display);
		font-size: 0.65rem;
		letter-spacing: 0.06em;
		color: var(--gr-ink-muted);
	}
	.gr-ability-score {
		font-size: 1.15rem;
		font-weight: 700;
		color: var(--gr-ink);
	}
	.gr-ability-mod {
		color: var(--gr-ink-muted);
		font-size: 0.85rem;
	}
	.gr-details {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		margin: 0.75rem 0;
		font-size: 0.9rem;
	}
	.gr-details div {
		display: flex;
		gap: 0.4rem;
	}
	.gr-details dt {
		font-family: var(--gr-font-display);
		font-size: 0.7rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--gr-ink-muted);
		min-width: 9rem;
	}
	.gr-details dd {
		margin: 0;
	}
	.gr-features {
		margin-top: 1rem;
	}
	.gr-features h3 {
		font-family: var(--gr-font-display);
		font-size: 0.8rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--gr-ink);
		border-bottom: 1px solid var(--gr-parchment-border);
		padding-bottom: 0.3rem;
		margin: 0 0 0.5rem;
	}
	.gr-action {
		margin: 0.4rem 0;
		font-size: 0.85rem;
		line-height: 1.45;
	}
	.gr-action b {
		font-style: italic;
	}
	.gr-action-recharge {
		background: var(--gr-parchment-highlight);
		border-left: 3px solid var(--gr-accent);
		border-radius: 0 8px 8px 0;
		padding: 0.5rem 0.75rem;
		margin: 0.5rem 0;
	}
	.gr-action-recharge p {
		margin: 0.2rem 0;
	}
	.gr-recharge-badge {
		font-family: var(--gr-font-display);
		font-size: 0.625rem;
		letter-spacing: 0.04em;
		color: var(--gr-accent);
		border: 1px solid var(--gr-cream-soft);
		border-radius: var(--gr-radius-sm);
		padding: 0.0625rem 0.375rem;
	}
	.gr-source {
		margin-top: 1rem;
		color: var(--gr-ink-muted);
		font-size: 0.8rem;
	}
	.gr-footer-actions {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
		margin-top: 1.25rem;
	}
	.gr-pill-btn {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--gr-ink-muted);
		border: 1.3px solid var(--gr-parchment-border-strong);
		border-radius: var(--gr-radius-md);
		padding: 0.375rem 0.75rem;
		background: none;
		text-decoration: none;
		cursor: pointer;
	}
	.gr-pill-btn-filled {
		background: var(--gr-accent);
		border-color: var(--gr-accent);
		color: white;
	}
	.gr-notes {
		display: block;
		margin-top: 0.75rem;
		font-family: var(--gr-font-display);
		font-size: 0.7rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--gr-ink-muted);
	}
	.gr-notes textarea {
		display: block;
		width: 100%;
		min-height: 4rem;
		margin-top: 0.4rem;
		padding: 0.5rem;
		font-family: var(--gr-font-body);
		font-size: 0.85rem;
		text-transform: none;
		letter-spacing: normal;
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-md);
	}
</style>
