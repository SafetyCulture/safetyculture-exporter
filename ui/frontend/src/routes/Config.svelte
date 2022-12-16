<script>
	import './common.css';
	import Select from 'svelte-select';

	const statusItems = [
		{value: "true", label: "Completed only"},
		{value: "false", label: "Incompleted only"},
		{value: "both", label: "Both - completed and incompleted"}
	];

	const archivedItems = [
		{value: "true", label: "Archived Only"},
		{value: "false", label: "Unarchived Only"},
		{value: "both", label: "Both - archived and unarchived"}
	];

	const dataExportFormatItems = [
		{value: "csv", label: "CSV"},
		{value: "sql", label: "SQL"}
	];

	const reportFormatItems = [
		{value: "PDF", label: "PDF"},
		{value: "WORD", label: "Word"},
		{value: "both", label: "Both - PDF and Word"},
	];

	const timezoneItems = [
		{value: "utc", label: "UTC"}
	];

	let templateCount = "N/A";
	let templateNames = ["placeholder text"];
	let dataExportFormat = "csv";

	function handleDataExport(event) {
		dataExportFormat = event.detail.value;
	}
</script>

<div class="config-page p-48">
	<section class="top-nav">
		<div class="nav-left">
			<div class="arrow-left">
				<img src="../images/arrow-left.png" alt="back arrow icon" width="15" height="15">
			</div>
			<div class="h1">Export Configuration</div>
		</div>
		<div class="nav-right">
			<button class="button button-white border-round-12">Save and Close</button>
			<button class="button button-purple m-left-8 border-round-12">Save and Export</button>
		</div>
	</section>
	<div class="config-body m-top-8">
		<section class="filters">
			<div class="filter-title">
				<div class="h3">Filters</div>
				<div class="text-weak m-top-8">Select which sets of data you want to export from your organization.</div>
			</div>
			<div class="label">Select templates</div>
				<div class="button-long selector border-weak border-round-8">
					<div class="templates">{templateNames}</div>
					<div class="template-button-right">
						<div class="count">{templateCount}</div>
						<img class="m-left-8" src="../images/arrow-right-compact.png" alt="right arrow icon" width="4" height="8">
					</div>
				</div>
			<div class="label">Date range</div>
			<div class="sub-label text-weak">From:</div>
			<div class="button-long selector border-weak border-round-8">
				<div>Date Picker(Unimplemented)</div>
				<img src="../images/calendar.png" alt="calendar icon" width="15" height="15">
			</div>
			<div class="label">Include inspections with the following status:</div>
			<div class="border-weak border-round-8 m-top-4">
				<Select
					items={statusItems}
					isClearable={false}
					>
				</Select>
			</div>
			<div class="label">Include archived inspections?</div>
			<div class="border-weak border-round-8 m-top-4">
				<Select
					items={archivedItems}
					isClearable={false}
				>
				</Select>
			</div>
		</section>
		<section class="export-details border-round-8">
			<div class="h3">Export details</div>
			<div class="label">Data export format</div>
			<div class="border-weak border-round-8 m-top-4">
				<Select
					items={dataExportFormatItems}
					isClearable={false}
					on:select={handleDataExport}
				>
				</Select>
			</div>
			{#if dataExportFormat === "sql"}
				<div>
					<div class="label">Database details:</div>
					<div class="sub-label text-weak">Host Address</div>
					<input class="input" type="text">
					<div class="sub-label text-weak">Host Port</div>
					<input class="input" type="text">
					<div class="sub-label text-weak">Username</div>
					<input class="input" type="text">
					<div class="sub-label text-weak">Password</div>
					<input class="input" type="password">
					<div class="sub-label text-weak">Name</div>
					<input class="input" type="text">
					<hr>
				</div>
			{/if}
			<div class="label">Report format</div>
			<Select
				items={reportFormatItems}
				isClearable={false}
			>
			</Select>
			<div class="folder-title">
				<div class="label">Folder location</div>
				<div class="link text-size-small">Want to change location?</div>
			</div>
			<div class="button-long selector border-weak border-round-8">
				<div class="text-weak">folder(unimplemented)</div>
				<img src="../images/folder.png" alt="folder icon" width="15" height="15">
			</div>
			<div class="label">Export timezone</div>
			<Select
				items={timezoneItems}
				isClearable={false}
			>
			</Select>
			<div class="label">Include:</div>
			<input type="checkbox" id="media" name="media" value="media">
			<label class="text-size-medium" for="media">Media</label>
		</section>
	</div>
</div>

<style>
	.nav-left .arrow-left {
		padding: 8px;
	}

	.config-body {
		display: flex;
		justify-content: space-between;
	}

	.filters {
		width: 55%;
	}

	.template-button-right {
		display: flex;
		align-items: center;
	}

	.count {
		background: #E5FAFF;
		border-radius: 100px;
		padding: 2px 10px;
		color: #0D75B5;
	}

	.export-details {
		width: 380px;
		height: 600px;
		background-color: #E9EEF6;
		padding: 16px;
		overflow-y: auto;
	}

	.folder-title {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
	}
</style>
