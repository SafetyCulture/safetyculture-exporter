<script>
	import './config.css';
</script>

<div class="config-page">
	<section class="top-nav">
		<div class="top-nav-button-back">Back</div>
		<div class="top-nav-header">Export Configuration</div>
		<div class="top-nav-button-save">Save and Close</div>
		<div class="top-nav-button-export">Save and Export</div>
	</section>
	<div class="config-body">
		<section class="filters">
			<div class="filter-title">
				<div class="filter-title-heading">Filters</div>
				<div class="filter-title-description">Select which sets of data you want to export from your organization.</div>
			</div>
			<div class="filter-templates">
				<div class="filter-templates-heading">Select templates</div>
				<div class="filter-templates-selector"></div>
			</div>
			<div class="filter-sites">
				<div class="filter-sites-heading">Select sites</div>
				<div class="filter-sites-selector"></div>
			</div>
			<div class="filter-dates">
				<div class="filter-dates-heading">Select sites</div>
				<div class="filter-dates-from">
					<div class="filter-dates-from-text">From:</div>
					<div class="filter-dates-from-selector"></div>
				</div>
				<div class="filter-dates-to">
					<div class="filter-dates-to-text">To:</div>
					<div class="filter-dates-to-selector"></div>
				</div>
			</div>
			<div class="filter-inspection-status">
				<div class="filter-inspection-status-heading">Include inspections with the following status:</div>
				<div class="filter-inspection-status-selector"></div>
			</div>
			<div class="filter-inspection-archived">
				<div class="filter-inspection-archived-heading">Include archived inspections?</div>
				<div class="filter-inspection-archived-selector"></div>
			</div>
		</section>
		<section class="export-details">
			<div class="export-details-heading">Export details</div>
			<div class="export-format">
				<div class="export-format-heading">Data export format</div>
				<div class="export-format-selector"></div>
			</div>
			<div class="report-format">
				<div class="report-format-heading">Report format</div>
				<div class="report-format-selector"></div>
			</div>
			<div class="folder-location">
				<div class="folder-location-heading">Folder location</div>
				<div class="folder-location-link">Want to change location?</div>
				<div class="folder-location-selector"></div>
			</div>
			<div class="export-timezone">
				<div class="export-timezone-heading">Export timezone</div>
				<div class="export-timezone-selector"></div>
			</div>
			<div class="export-inclusions">
				<div class="export-inclusions-heading">Include:</div>
				<div class="export-inclusions-selector"></div>
			</div>
		</section>
	</div>
</div>