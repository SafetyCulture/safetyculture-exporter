<script>
	import './common.css';
	import './config.css';
	import DatePicker from "../components/DatePicker.svelte";
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
	<div class="config-body">
		<section class="filters">
			<div class="filter-title">
				<div class="h3">Filters</div>
				<div class="text-weak m-top-8">Select which sets of data you want to export from your organization.</div>
			</div>
			<div class="label">Select templates</div>
				<div class="template-button selector border-weak border-round-8">
					<div class="templates">Surface Diamond Drill Safety Weekly Inspection Form</div>
					<div class="template-button-right">
						<div class="count">10</div>
						<img src="../images/arrow-right-compact.png" alt="right arrow icon" width="10" height="10">
					</div>
				</div>
			<div class="label">Inspection date range</div>
			<div class="sub-label text-weak">From:</div>
			<DatePicker></DatePicker>
			<div class="sub-label text-weak">To:</div>
			<DatePicker></DatePicker>
			<div class="label">Include inspections with the following status:</div>
			<select name="status">
				<option value="true">complete</option>
				<option value="false">incomplete</option>
				<option value="both">both</option>
			</select>
			<div class="label">Include archived inspections?</div>
			<select name="status">
				<option value="true">complete</option>
				<option value="false">incomplete</option>
				<option value="both">both</option>
			</select>
		</section>
		<section class="export-details">
			<div class="h3">Export details</div>
			<div class="label">Data export format</div>
			<select name="export-format">
				<option value="csv">csv</option>
				<option value="sql">sql</option>
			</select>
			<div class="label">Report format</div>
			<select name="report-format">
				<option value="pdf">PDF</option>
				<option value="word">Word</option>
				<option value="none">Don't export reports</option>
			</select>
			<div class="label">Folder location</div>
			<div class="link">Want to change location?</div>
			<div class="folder-selector">
				<div>text</div>
				<div>icon</div>
			</div>
			<div class="label">Export timezone</div>
			<select name="export-format">
				<option value="utc">utc</option>
				<option value="1">utc+1</option>
			</select>
			<div class="label">Include:</div>
			<input type="checkbox" id="media" name="media" value="media">
			<label for="media">Media</label>
		</section>
	</div>
</div>