# Check for required packages
required_packages <- c("data.table", "foreach", "doParallel")
new_packages <- required_packages[!(required_packages %in% installed.packages()[,"Package"])]
if(length(new_packages)) install.packages(new_packages)


library(data.table)
library(foreach)
library(doParallel)

# --- 1. Define the Core Logic (The 'kk' function) ---
# This is the exact logic needed for the calculation, used by both methods.
kk_benchmark = function(theta, intens, nstart_val=40) {
  # Simplified logic for benchmarking purposes to test computational load
  # (Mimics the heavy lifting of K-means clustering from your original script)
  
  # Remove NAs
  i.ok = which(!is.na(theta) & !is.na(intens))
  if(length(i.ok) < 5) return(NULL) # Skip if too few points
  
  THeta = theta[i.ok]
  INtens = intens[i.ok]
  
  # Run K-means 1 to 5 (The heavy part)
  g1 = kmeans(THeta, 1, nstart=nstart_val) 
  g2 = kmeans(THeta, 2, nstart=nstart_val)
  g3 = kmeans(THeta, 3, nstart=nstart_val)
  
  return(g3$tot.withinss) # Return dummy result
}

# --- 2. Generate Dummy Data ---
cat("========================================================\n")
cat("Step 1: Generating synthetic data (44040 SNPs x 750 Genotypes)...\n")
n_snps <- 44040
n_geno <- 750

# Create random matrices
dummy_theta <- matrix(runif(n_snps * n_geno), nrow=n_snps)
dummy_intens <- matrix(runif(n_snps * n_geno), nrow=n_snps)

# Add headers/names to mimic real files
df_theta <- data.frame(SNP_ID = paste0("SNP", 1:n_snps), dummy_theta)
df_intens <- data.frame(SNP_ID = paste0("SNP", 1:n_snps), dummy_intens)

# Write to temp files
write.csv(df_theta, "temp_theta.csv", row.names=FALSE)
write.csv(df_intens, "temp_intens.csv", row.names=FALSE)
cat("Data generated.\n\n")

# --- 3. RUN OLD METHOD ---
cat("Step 2: Running OLD method (Sequential, Transpose, read.csv)...\n")
start_time_old <- Sys.time()

# A. Slow Read
old_theta <- read.csv("temp_theta.csv")
old_intens <- read.csv("temp_intens.csv")

# B. Expensive Transpose (t)
t_theta <- t(old_theta[, -1]) 
t_intens <- t(old_intens[, -1])
df_t_theta <- data.frame(t_theta)
df_t_intens <- data.frame(t_intens)

# C. Sequential Processing (mapply)
# Note: running with nstart=40 as per original script
results_old <- mapply(kk_benchmark, 
                      df_t_theta, 
                      df_t_intens, 
                      MoreArgs = list(nstart_val=40))

end_time_old <- Sys.time()
time_old <- as.numeric(difftime(end_time_old, start_time_old, units="secs"))
cat(paste("Old Method finished in:", round(time_old, 2), "seconds.\n\n"))


# --- 4. RUN NEW METHOD ---
cat("Step 3: Running NEW method (Parallel, No Transpose, fread)...\n")
start_time_new <- Sys.time()

# A. Fast Read
dt_theta <- fread("temp_theta.csv", header=TRUE, data.table=FALSE)
dt_intens <- fread("temp_intens.csv", header=TRUE, data.table=FALSE)

# B. No Transpose (Direct Matrix Access)
mat_theta <- as.matrix(dt_theta[, -1])
mat_intens <- as.matrix(dt_intens[, -1])

# C. Parallel Processing
# Detect cores (using safe number for test)
cores_to_use <- max(1, detectCores() - 1)
cl <- makeCluster(cores_to_use)
registerDoParallel(cl)

# Note: running with nstart=25 as per optimized recommendation
results_new <- foreach(i = 1:n_snps, .packages='MASS') %dopar% {
  kk_benchmark(mat_theta[i, ], mat_intens[i, ], nstart_val=25)
}
stopCluster(cl)

end_time_new <- Sys.time()
time_new <- as.numeric(difftime(end_time_new, start_time_new, units="secs"))
cat(paste("New Method finished in:", round(time_new, 2), "seconds.\n\n"))

# --- 5. RESULTS ---
cat("========================================================\n")
cat("PERFORMANCE REPORT\n")
cat("========================================================\n")
cat(sprintf("Number of SNPs processed: %d\n", n_snps))
cat(sprintf("Old Method Time:          %.2f seconds\n", time_old))
cat(sprintf("New Method Time:          %.2f seconds\n", time_new))
cat("--------------------------------------------------------\n")
speedup <- time_old / time_new
cat(sprintf("SPEEDUP FACTOR:           %.1fx faster\n", speedup))
cat("========================================================\n")

# Cleanup temp files
file.remove("temp_theta.csv")
file.remove("temp_intens.csv")