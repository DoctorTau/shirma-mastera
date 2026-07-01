<script lang="ts">
	import type { StatBlock, Feature } from '$lib/types';
	import '$lib/styles/grifel.css';

	let { statblock = $bindable() }: { statblock: StatBlock } = $props();

	const abilityFields: [keyof StatBlock['abilities'], string][] = [
		['str', 'СИЛ'],
		['dex', 'ЛОВ'],
		['con', 'ТЕЛ'],
		['int', 'ИНТ'],
		['wis', 'МДР'],
		['cha', 'ХАР']
	];

	function abilityMod(score: number): number {
		return Math.floor((score - 10) / 2);
	}

	function onScoreInput(key: keyof StatBlock['abilities'], value: number) {
		statblock.abilities[key] = { score: value, mod: abilityMod(value) };
	}

	const featureLists: { key: keyof StatBlock; label: string }[] = [
		{ key: 'traits', label: 'Особенности' },
		{ key: 'actions', label: 'Действия' },
		{ key: 'bonusActions', label: 'Бонусные действия' },
		{ key: 'reactions', label: 'Реакции' },
		{ key: 'legendaryActions', label: 'Легендарные действия' }
	];

	function addFeature(key: keyof StatBlock) {
		const list = (statblock[key] as Feature[] | undefined) ?? [];
		(statblock as unknown as Record<string, Feature[]>)[key] = [...list, { name: '', description: '' }];
	}

	function removeFeature(key: keyof StatBlock, index: number) {
		const list = (statblock[key] as Feature[] | undefined) ?? [];
		(statblock as unknown as Record<string, Feature[]>)[key] = list.filter((_, i) => i !== index);
	}
</script>

<div class="editor gr-editor">
	<div class="row">
		<label>Размер<input bind:value={statblock.sizeRu} /></label>
		<label>Тип<input bind:value={statblock.type} /></label>
		<label>Мировоззрение<input bind:value={statblock.alignment} /></label>
	</div>
	<div class="row">
		<label>КД<input bind:value={statblock.armorClass} /></label>
		<label>Источник КД<input bind:value={statblock.armorSource} /></label>
		<label>Хиты<input bind:value={statblock.hitPoints} /></label>
		<label>Кость хитов<input bind:value={statblock.hitDice} /></label>
	</div>

	<fieldset>
		<legend>Характеристики</legend>
		<div class="abilities-row">
			{#each abilityFields as [key, label] (key)}
				<label class="ability-input">
					{label}
					<input
						type="number"
						value={statblock.abilities[key]?.score ?? 10}
						oninput={(e) => onScoreInput(key, Number((e.target as HTMLInputElement).value))}
					/>
					<span class="mod">{abilityMod(statblock.abilities[key]?.score ?? 10) >= 0 ? '+' : ''}{abilityMod(statblock.abilities[key]?.score ?? 10)}</span>
				</label>
			{/each}
		</div>
	</fieldset>

	<div class="row">
		<label>Чувства<input bind:value={statblock.senses} /></label>
		<label
			>Пасс. внимательность<input type="number" bind:value={statblock.passivePerception} /></label
		>
		<label>Языки<input bind:value={statblock.languages} /></label>
	</div>
	<div class="row">
		<label>Опасность (CR)<input bind:value={statblock.challengeRating} /></label>
		<label>Опыт<input type="number" bind:value={statblock.experiencePoints} /></label>
		<label>Бонус мастерства<input type="number" bind:value={statblock.proficiencyBonus} /></label>
	</div>
	<div class="row">
		<label class="grow">Уязвимости (через запятую)
			<input
				value={(statblock.vulnerabilities ?? []).join(', ')}
				oninput={(e) => (statblock.vulnerabilities = (e.target as HTMLInputElement).value.split(',').map((s) => s.trim()).filter(Boolean))}
			/>
		</label>
		<label class="grow">Сопротивления
			<input
				value={(statblock.resistances ?? []).join(', ')}
				oninput={(e) => (statblock.resistances = (e.target as HTMLInputElement).value.split(',').map((s) => s.trim()).filter(Boolean))}
			/>
		</label>
		<label class="grow">Иммунитеты
			<input
				value={(statblock.immunities ?? []).join(', ')}
				oninput={(e) => (statblock.immunities = (e.target as HTMLInputElement).value.split(',').map((s) => s.trim()).filter(Boolean))}
			/>
		</label>
	</div>

	{#each featureLists as { key, label } (key)}
		<fieldset>
			<legend>{label}</legend>
			{#each (statblock[key] as Feature[] | undefined) ?? [] as feature, i (i)}
				<div class="feature-row">
					<input placeholder="Название" bind:value={feature.name} />
					<textarea placeholder="Описание" bind:value={feature.description}></textarea>
					<button type="button" onclick={() => removeFeature(key, i)}>Удалить</button>
				</div>
			{/each}
			<button type="button" onclick={() => addFeature(key)}>+ Добавить</button>
		</fieldset>
	{/each}
</div>

<style>
	.gr-editor {
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-md);
		font-family: var(--gr-font-body);
	}
	.row {
		display: flex;
		gap: var(--gr-space-md);
		flex-wrap: wrap;
	}
	label {
		display: flex;
		flex-direction: column;
		font-size: 0.8rem;
		color: var(--gr-ink-muted);
		gap: 0.25rem;
	}
	label.grow {
		flex: 1;
	}
	input,
	textarea {
		font-family: var(--gr-font-body);
		padding: 0.45rem 0.6rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-sm);
		font-size: 0.9rem;
	}
	textarea {
		min-height: 3rem;
		resize: vertical;
	}
	fieldset {
		border: 1.5px solid var(--gr-parchment-border);
		border-radius: var(--gr-radius-md);
		padding: 0.75rem;
	}
	legend {
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--gr-ink-muted);
		padding: 0 0.4rem;
	}
	.abilities-row {
		display: grid;
		grid-template-columns: repeat(6, 1fr);
		gap: var(--gr-space-md);
	}
	.ability-input {
		align-items: center;
		min-width: 0;
	}
	.ability-input input {
		width: 100%;
		text-align: center;
	}
	@media (max-width: 680px) {
		.abilities-row {
			grid-template-columns: repeat(3, 1fr);
		}
	}
	.mod {
		color: var(--gr-ink-faint);
	}
	.feature-row {
		display: flex;
		gap: var(--gr-space-sm);
		margin-bottom: var(--gr-space-sm);
		align-items: flex-start;
	}
	.feature-row input {
		flex: 0 0 10rem;
	}
	.feature-row textarea {
		flex: 1;
	}
	button {
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		padding: 0.4rem 0.8rem;
		border-radius: var(--gr-radius-sm);
		border: 1.3px solid var(--gr-parchment-border-strong);
		background: var(--gr-parchment-card);
		color: var(--gr-ink-muted);
		cursor: pointer;
	}
</style>
